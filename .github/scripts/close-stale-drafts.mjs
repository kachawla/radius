// @ts-check

import { createHash } from "node:crypto";

const DAY = 24 * 60 * 60 * 1000;
// GitHub's comment, label-event and PR clocks can differ by a second or two.
const CLOCK_SKEW = 15 * 1000;
const LABEL = "stale-draft";
const BOT = "github-actions[bot]";
const MESSAGE =
  "This draft pull request has been inactive for 7 days and is marked as stale. Comment, push an update, remove the `stale-draft` label, or mark it ready for review to keep it open. If there is no further activity, it will be closed after 1 day. Feel free to reopen it when you are ready to continue.";

function timestamp(value) {
  const result = Date.parse(value);
  if (!Number.isFinite(result)) {
    throw new Error(`Invalid GitHub timestamp: ${value}`);
  }
  return result;
}

function hasLabel(pr, name) {
  return pr.labels.some((label) => label.name === name);
}

function warningBody(pr) {
  const content = createHash("sha256")
    .update(JSON.stringify([pr.title, pr.body]))
    .digest("hex");
  return `${MESSAGE}\n\n<!-- stale-draft:v1 ${pr.head.sha} ${content} -->`;
}

function unchanged(before, after) {
  return (
    before.state === after.state &&
    before.draft === after.draft &&
    before.locked === after.locked &&
    before.updated_at === after.updated_at &&
    before.head.sha === after.head.sha &&
    before.title === after.title &&
    before.body === after.body &&
    hasLabel(before, "stale") === hasLabel(after, "stale") &&
    hasLabel(before, LABEL) === hasLabel(after, LABEL)
  );
}

/** Reconcile only drafts and draft-owned labels. All mutations disable HTTP retries. */
export default async function closeStaleDrafts({
  github,
  context,
  core,
  now = Date.now()
}) {
  const repo = context.repo;
  // Collect the inventory before writes, which change updated_at and pagination order.
  const prs = await github.paginate(github.rest.pulls.list, {
    ...repo,
    state: "open",
    sort: "created",
    direction: "asc",
    per_page: 100
  });
  for (const candidate of prs) {
    if (!candidate.draft && !hasLabel(candidate, LABEL)) continue;
    const pull = { ...repo, pull_number: candidate.number };
    const issue = { ...repo, issue_number: candidate.number };
    const read = async () => (await github.rest.pulls.get(pull)).data;
    const pr = await read();
    if (pr.state !== "open" || pr.locked) continue;

    const removeLabel = async (name = LABEL) => {
      const current = await read();
      if (!unchanged(pr, current)) {
        core.info(
          `PR ${pr.number} changed during reconciliation; leaving it for the next run.`
        );
        return;
      }
      await github.rest.issues.removeLabel({
        ...issue,
        name,
        request: { retries: 0 }
      });
    };
    if (!pr.draft) {
      if (hasLabel(pr, LABEL)) await removeLabel();
      continue;
    }
    if (hasLabel(pr, "stale")) {
      // Drafts previously belonged to the shared 28-day policy.
      await removeLabel("stale");
      continue;
    }

    if (!hasLabel(pr, LABEL)) {
      if (now - timestamp(pr.updated_at) < 7 * DAY) continue;
      const current = await read();
      if (!unchanged(pr, current)) continue;
      // Adding the label first makes an interrupted warning restart with a fresh timer.
      await github.rest.issues.addLabels({
        ...issue,
        labels: [LABEL],
        request: { retries: 0 }
      });
      const labeled = await read();
      if (
        labeled.state !== "open" ||
        !labeled.draft ||
        labeled.locked ||
        !hasLabel(labeled, LABEL) ||
        warningBody(labeled) !== warningBody(pr)
      )
        continue;
      await github.rest.issues.createComment({
        ...issue,
        body: warningBody(labeled),
        request: { retries: 0 }
      });
      core.info(`Warned inactive draft PR ${pr.number}.`);
      continue;
    }

    const events = await github.paginate(
      github.rest.issues.listEventsForTimeline,
      {
        ...issue,
        per_page: 100
      }
    );
    const comments = await github.paginate(github.rest.issues.listComments, {
      ...issue,
      per_page: 100
    });
    const labelEvents = events.filter(
      (event) => event.event === "labeled" && event.label?.name === LABEL
    );
    if (!labelEvents.length) {
      throw new Error(
        `PR ${pr.number}: stale-draft label event is not visible; retry the workflow.`
      );
    }
    const labeledAt = Math.max(
      0,
      ...labelEvents.map((event) => timestamp(event.created_at))
    );
    const warnings = comments.filter(
      (comment) =>
        comment.user?.login === BOT &&
        comment.user?.type === "Bot" &&
        comment.body === warningBody(pr) &&
        comment.created_at === comment.updated_at &&
        timestamp(comment.created_at) + CLOCK_SKEW >= labeledAt
    );
    const warning = warnings.sort((a, b) => b.id - a.id)[0];
    if (!warning) {
      // Missing/deleted/edited warnings and partial writes can never authorize closure.
      await removeLabel();
      continue;
    }
    const warnedAt = timestamp(warning.created_at);
    const ownLabelEvents = labelEvents.filter(
      (event) =>
        event.actor?.login === BOT &&
        event.actor?.type === "Bot" &&
        timestamp(event.created_at) <= warnedAt + CLOCK_SKEW
    );
    const activity =
      events.some((event) => {
        if (event.event === "commented" && event.id === warning.id)
          return false;
        if (ownLabelEvents.some((label) => label.id === event.id)) return false;
        // These indirect events do not update the PR's inactivity clock.
        if (
          ["committed", "referenced", "cross-referenced"].includes(
            event.event
          ) ||
          event.event.includes("project")
        )
          return false;
        if (event.event === "reviewed") {
          if (!event.submitted_at) {
            core.warning(
              `PR ${pr.number}: review timestamp is missing; cancelling draft closure.`
            );
            return true;
          }
          return timestamp(event.submitted_at) >= warnedAt;
        }
        if (event.event === "line-commented") {
          if (!Array.isArray(event.comments)) {
            core.warning(
              `PR ${pr.number}: review comments are missing; cancelling draft closure.`
            );
            return true;
          }
          return event.comments.some(
            (comment) => timestamp(comment.updated_at) >= warnedAt
          );
        }
        return event.created_at && timestamp(event.created_at) >= warnedAt;
      }) ||
      comments.some(
        (comment) =>
          comment.id !== warning.id && timestamp(comment.updated_at) >= warnedAt
      ) ||
      timestamp(pr.updated_at) > warnedAt + CLOCK_SKEW;
    if (activity) {
      await removeLabel();
      continue;
    }
    if (now - warnedAt < DAY) continue;
    const reviewComments = await github.paginate(
      github.rest.pulls.listReviewComments,
      { ...pull, per_page: 100 }
    );
    if (
      reviewComments.some(
        (comment) => timestamp(comment.updated_at) >= warnedAt
      )
    ) {
      await removeLabel();
      continue;
    }
    const current = await read();
    if (!unchanged(pr, current)) {
      core.info(`PR ${pr.number} changed before closure; leaving it open.`);
      continue;
    }
    await github.rest.pulls.update({
      ...pull,
      state: "closed",
      request: { retries: 0 }
    });
    core.info(`Closed inactive draft PR ${pr.number}.`);
  }
}
