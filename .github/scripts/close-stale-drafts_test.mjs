import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import test from "node:test";
import run from "./close-stale-drafts.mjs";

const DAY = 86_400_000;
const START = Date.parse("2026-10-01T12:00:00Z");
const iso = (time) => new Date(time).toISOString();
const bot = { login: "github-actions[bot]", type: "Bot" };
const human = { login: "contributor", type: "User" };
const LABEL = "stale-draft";

// REST rules: pulls.list defaults to open, pages max at 100; issue comments are
// ascending by ID. Label deletion returns 404 when absent. Updating a PR has no
// atomic draft/head precondition. Timelines contain heterogeneous event shapes.
// https://docs.github.com/en/rest/pulls/pulls
// https://docs.github.com/en/rest/issues/comments
// https://docs.github.com/en/rest/issues/labels
// https://docs.github.com/en/rest/issues/timeline
function fixture(count = 1) {
  let time = START;
  let nextId = 1;
  const prs = Array.from({ length: count }, (_, index) => ({
    number: index + 1,
    state: "open",
    draft: true,
    locked: false,
    title: "Work in progress",
    body: "Draft implementation",
    updated_at: iso(START - 7 * DAY),
    head: { sha: "a".repeat(40) },
    labels: []
  }));
  const comments = new Map(prs.map((pr) => [pr.number, []]));
  const reviewComments = new Map(prs.map((pr) => [pr.number, []]));
  const events = new Map(prs.map((pr) => [pr.number, []]));
  const writes = [];
  const reads = [];
  let hook = () => {};
  const fail = new Map();
  const methods = {};
  function method(name, fn) {
    methods[name] = async (args) => {
      hook(name, args);
      if (fail.has(name)) throw fail.get(name);
      return { data: structuredClone(fn(args)) };
    };
    return methods[name];
  }
  const prFor = (args) => {
    assert.equal(args.owner, "owner");
    assert.equal(args.repo, "repo");
    const pr = prs.find(
      (pr) => pr.number === (args.pull_number ?? args.issue_number)
    );
    assert.ok(pr, "PR must exist");
    return pr;
  };
  function write(name, args) {
    assert.deepEqual(args.request, { retries: 0 });
    writes.push(name);
  }
  function event(pr, data) {
    const result = {
      id: nextId++,
      created_at: iso(time),
      actor: human,
      ...data
    };
    if (["reviewed", "line-commented", "committed"].includes(result.event)) {
      delete result.created_at;
    }
    events.get(pr.number).push(result);
    pr.updated_at = iso(time);
    return result;
  }
  const pulls = {
    listReviewComments: method("review-comments", (args) =>
      reviewComments.get(prFor(args).number)
    ),
    list: method("list", (args) => {
      assert.equal(args.state ?? "open", "open");
      assert.equal(args.sort, "created");
      assert.equal(args.direction, "asc");
      return prs.filter((pr) => pr.state === "open");
    }),
    get: method("get", (args) => {
      reads.push(args.pull_number);
      return prFor(args);
    }),
    update: method("close", (args) => {
      assert.equal(args.state, "closed");
      const pr = prFor(args);
      assert.equal(pr.state, "open");
      write("close", args);
      pr.state = "closed";
      return pr;
    })
  };
  const issues = {
    listComments: method("comments", (args) =>
      comments.get(prFor(args).number)
    ),
    listEventsForTimeline: method("events", (args) =>
      events.get(prFor(args).number)
    ),
    addLabels: method("label", (args) => {
      const pr = prFor(args);
      assert.deepEqual(args.labels, [LABEL]);
      write("label", args);
      // The label is provisioned by the repository's label-sync workflow.
      if (!pr.labels.some((label) => label.name === LABEL)) {
        pr.labels.push({ name: LABEL });
        event(pr, { event: "labeled", label: { name: LABEL }, actor: bot });
      }
      return pr.labels;
    }),
    removeLabel: method("unlabel", (args) => {
      const pr = prFor(args);
      assert.ok([LABEL, "stale"].includes(args.name));
      if (!pr.labels.some((label) => label.name === args.name)) {
        throw Object.assign(new Error("Label does not exist"), { status: 404 });
      }
      write("unlabel", args);
      pr.labels = pr.labels.filter((label) => label.name !== args.name);
      event(pr, { event: "unlabeled", label: { name: args.name }, actor: bot });
      return pr.labels;
    }),
    createComment: method("warn", (args) => {
      const pr = prFor(args);
      assert.equal(typeof args.body, "string");
      write("warn", args);
      const comment = {
        id: nextId++,
        body: args.body,
        user: bot,
        created_at: iso(time),
        updated_at: iso(time)
      };
      comments.get(pr.number).push(comment);
      event(pr, { ...comment, event: "commented" });
      return comment;
    })
  };
  const github = {
    rest: { pulls, issues },
    paginate: async (method, args) => {
      assert.equal(args.per_page, 100);
      const { data } = await method(args);
      // Deliberately span multiple pages; mutations must not happen mid-inventory.
      const pages = [];
      for (let offset = 0; offset < data.length; offset += args.per_page) {
        pages.push(data.slice(offset, offset + args.per_page));
      }
      return pages.flat();
    }
  };
  return {
    github,
    reviewComments,
    prs,
    comments,
    events,
    writes,
    reads,
    fail,
    event,
    setTime: (value) => {
      time = value;
    },
    hook: (fn) => {
      hook = fn;
    },
    run: () =>
      run({
        github,
        context: { repo: { owner: "owner", repo: "repo" } },
        core: { info() {}, warning() {} },
        now: time
      })
  };
}

test("warns exactly at seven days, waits a full day, closes without touching branches", async () => {
  const f = fixture();
  f.setTime(START - 1);
  await f.run();
  assert.deepEqual(f.writes, []);
  f.setTime(START);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn"]);
  const body = f.comments.get(1)[0].body;
  assert.match(body, /inactive for 7 days/);
  assert.match(body, /closed after 1 day/);
  f.setTime(START + DAY - 1);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn"]);
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "close"]);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "close"]);
});

test("leaves review-ready PRs and their existing stale labels alone", async () => {
  const f = fixture();
  Object.assign(f.prs[0], { draft: false, labels: [{ name: "stale" }] });
  await f.run();
  assert.deepEqual(f.writes, []);
  assert.deepEqual(f.reads, []);
});

for (const state of ["closed", "locked"]) {
  test(`does not mutate ${state} PRs`, async () => {
    const f = fixture();
    if (state === "closed") f.prs[0].state = "closed";
    else f.prs[0].locked = true;
    await f.run();
    assert.deepEqual(f.writes, []);
  });
}

for (const activity of [
  "comment",
  "edited-comment",
  "push",
  "review",
  "body-edit",
  "ready",
  "ready-and-draft",
  "remove-and-reapply-label",
  "reopen"
]) {
  test(`${activity} cancels pending draft closure`, async () => {
    const f = fixture();
    await f.run();
    f.setTime(START + 1000);
    const pr = f.prs[0];
    if (activity === "push") pr.head.sha = "b".repeat(40);
    if (activity === "body-edit") pr.body = "Updated description";
    if (activity === "ready") pr.draft = false;
    if (activity === "comment" || activity === "edited-comment") {
      f.comments.get(1).push({
        id: 100,
        user: human,
        body: "Still working",
        created_at: iso(START - DAY),
        updated_at: iso(START + 1000)
      });
    }
    if (activity === "ready-and-draft") {
      f.event(pr, { event: "ready_for_review" });
      f.event(pr, { event: "convert_to_draft" });
    }
    if (activity === "remove-and-reapply-label") {
      f.event(pr, { event: "unlabeled", label: { name: LABEL } });
      f.event(pr, { event: "labeled", label: { name: LABEL } });
    }
    if (activity === "reopen") f.event(pr, { event: "reopened" });
    if (activity === "review") {
      f.event(pr, { event: "reviewed", submitted_at: iso(START + 1000) });
    }
    pr.updated_at = iso(START + 1000);
    f.setTime(START + 2 * DAY);
    await f.run();
    assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
    assert.equal(pr.state, "open");
    assert.deepEqual(pr.labels, []);
    await f.run();
    assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
  });
}

test("same-second activity and ready/draft transitions invalidate the warning", async () => {
  const f = fixture();
  await f.run();
  f.event(f.prs[0], { event: "ready_for_review" });
  f.event(f.prs[0], { event: "convert_to_draft" });
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
});

test("bot warning and label clocks may differ from PR time by a few seconds", async () => {
  const f = fixture();
  await f.run();
  f.prs[0].updated_at = iso(START + 1000);
  f.events.get(1)[0].created_at = iso(START + 2000);
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "close"]);
});

for (const activity of ["title", "body", "review", "inline-comment"]) {
  test(`${activity} in the bot clock tolerance window still cancels closure`, async () => {
    const f = fixture();
    await f.run();
    const pr = f.prs[0];
    if (activity === "title") pr.title = "Updated title";
    if (activity === "body") pr.body = "Updated body";
    f.setTime(START + 1000);
    if (activity === "review") {
      f.event(pr, { event: "reviewed", submitted_at: iso(START + 1000) });
    }
    if (activity === "inline-comment") {
      f.event(pr, {
        event: "line-commented",
        comments: [
          { created_at: iso(START - DAY), updated_at: iso(START + 1000) }
        ]
      });
    }
    pr.updated_at = iso(START + 1000);
    f.setTime(START + DAY);
    await f.run();
    assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
  });
}

test("indirect references and project events do not cancel inactivity", async () => {
  const f = fixture();
  await f.run();
  f.setTime(START + 1000);
  for (const event of [
    "cross-referenced",
    "referenced",
    "added_to_project_v2",
    "project_v2_item_status_changed"
  ]) {
    f.event(f.prs[0], { id: null, event });
  }
  f.prs[0].updated_at = iso(START);
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "close"]);
});

test("inline comment edits inside clock tolerance cancel closure even without timeline events", async () => {
  const f = fixture();
  await f.run();
  f.reviewComments.get(1).push({
    id: 101,
    body: "Updated review comment",
    user: human,
    created_at: iso(START - DAY),
    updated_at: iso(START + 1000)
  });
  f.prs[0].updated_at = iso(START + 1000);
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
});

for (const event of ["reviewed", "line-commented"]) {
  test(`incomplete ${event} metadata keeps that PR open without blocking other drafts`, async () => {
    const f = fixture(2);
    await f.run();
    f.event(f.prs[0], { event });
    f.setTime(START + DAY);
    await f.run();
    assert.equal(f.prs[0].state, "open");
    assert.equal(f.prs[1].state, "closed");
    assert.ok(!f.prs[0].labels.some((label) => label.name === LABEL));
  });
}

test("migrates the old stale label off drafts without touching ready PRs", async () => {
  const f = fixture(2);
  f.prs.forEach((pr) => {
    pr.labels = [{ name: "stale" }];
  });
  f.prs[1].draft = false;
  await f.run();
  assert.deepEqual(f.writes, ["unlabel"]);
  assert.deepEqual(f.prs[0].labels, []);
  assert.deepEqual(f.prs[1].labels, [{ name: "stale" }]);
});

for (const damage of [
  "missing",
  "edited",
  "spoofed-user",
  "spoofed-bot-type"
]) {
  test(`${damage} warning cannot authorize closure`, async () => {
    const f = fixture();
    await f.run();
    const warning = f.comments.get(1)[0];
    if (damage === "missing") f.comments.set(1, []);
    if (damage === "edited") warning.updated_at = iso(START + 1000);
    if (damage === "spoofed-user") warning.user = human;
    if (damage === "spoofed-bot-type") warning.user = { ...bot, type: "User" };
    f.setTime(START + DAY);
    await f.run();
    assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
  });
}

test("partial warning failure never closes and retries with a new inactivity interval", async () => {
  const f = fixture();
  f.fail.set("warn", new Error("Network timeout"));
  await assert.rejects(f.run(), /Network timeout/);
  f.fail.clear();
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "unlabel"]);
  f.setTime(START + 8 * DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "unlabel", "label", "warn"]);
  await f.run();
  assert.equal(f.prs[0].state, "open");
});

test("a successful but uncertain warning is discovered without duplicating it", async () => {
  const f = fixture();
  await f.run();
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn"]);
});

test("duplicate warnings use the newest timer rather than closing early", async () => {
  const f = fixture();
  await f.run();
  const latest = {
    ...f.comments.get(1)[0],
    id: 100,
    created_at: iso(START + 1000),
    updated_at: iso(START + 1000)
  };
  f.comments.get(1).push(latest);
  f.prs[0].updated_at = latest.created_at;
  f.setTime(START + DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn"]);
  f.setTime(START + DAY + 1000);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "close"]);
});

test("missing timeline label visibility fails closed", async () => {
  const f = fixture();
  await f.run();
  f.events.set(1, []);
  f.setTime(START + DAY);
  await assert.rejects(f.run(), /label event is not visible/);
  assert.deepEqual(f.writes, ["label", "warn"]);
});

test("old warning cannot be reused after a later label application", async () => {
  const f = fixture();
  await f.run();
  f.setTime(START + DAY);
  f.event(f.prs[0], { event: "labeled", label: { name: LABEL } });
  f.setTime(START + 2 * DAY);
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn", "unlabel"]);
});

for (const change of [
  "ready",
  "comment",
  "push",
  "locked",
  "closed",
  "unlabel"
]) {
  test(`${change} just before closure leaves the PR open`, async () => {
    const f = fixture();
    await f.run();
    f.setTime(START + DAY);
    let gets = 0;
    f.hook((name) => {
      if (name !== "get" || ++gets !== 2) return;
      const pr = f.prs[0];
      if (change === "ready") pr.draft = false;
      if (change === "comment") pr.updated_at = iso(START + DAY);
      if (change === "push") pr.head.sha = "b".repeat(40);
      if (change === "locked") pr.locked = true;
      if (change === "closed") pr.state = "closed";
      if (change === "unlabel") pr.labels = [];
    });
    await f.run();
    assert.deepEqual(f.writes, ["label", "warn"]);
  });
}

test("activity just before labeling is not overwritten", async () => {
  const f = fixture();
  let gets = 0;
  f.hook((name) => {
    if (name === "get" && ++gets === 2) f.prs[0].updated_at = iso(START);
  });
  await f.run();
  assert.deepEqual(f.writes, []);
});

test("a draft becoming ready after labeling is not warned", async () => {
  const f = fixture();
  f.hook((name) => {
    if (name === "get" && f.writes.includes("label")) f.prs[0].draft = false;
  });
  await f.run();
  assert.deepEqual(f.writes, ["label"]);
});

test("concurrent changes prevent removing a label from a different state", async () => {
  const f = fixture();
  await f.run();
  f.prs[0].draft = false;
  let gets = 0;
  f.hook((name) => {
    if (name === "get" && ++gets === 2) f.prs[0].draft = true;
  });
  await f.run();
  assert.deepEqual(f.writes, ["label", "warn"]);
});

test("paginates inventory, comments, and timeline before changing PR state", async () => {
  const f = fixture(101);
  await f.run();
  assert.equal(f.writes.filter((write) => write === "warn").length, 101);
  for (const pr of f.prs) {
    for (let index = 0; index < 101; index++) {
      f.event(pr, {
        event: "committed",
        committer: { date: iso(START - DAY) }
      });
      f.comments.get(pr.number).push({
        id: 1000 + index,
        body: "Old comment",
        user: human,
        created_at: iso(START - DAY),
        updated_at: iso(START - DAY)
      });
    }
    pr.updated_at = iso(START);
  }
  f.setTime(START + DAY);
  await f.run();
  assert.equal(f.writes.filter((write) => write === "close").length, 101);
});

test("API read/write errors propagate without further writes", async () => {
  for (const operation of [
    "list",
    "get",
    "label",
    "comments",
    "events",
    "unlabel",
    "close",
    "review-comments"
  ]) {
    const f = fixture();
    if (
      ["comments", "events", "unlabel", "close", "review-comments"].includes(
        operation
      )
    ) {
      await f.run();
      f.setTime(START + DAY);
      if (operation === "unlabel") f.prs[0].draft = false;
    }
    const writes = [...f.writes];
    f.fail.set(operation, new Error(`${operation} failed`));
    await assert.rejects(f.run(), new RegExp(`${operation} failed`));
    assert.deepEqual(f.writes, writes);
  }
});

test("invalid GitHub dates fail instead of authorizing a write", async () => {
  const f = fixture();
  f.prs[0].updated_at = "invalid";
  await assert.rejects(f.run(), /Invalid GitHub timestamp/);
  assert.deepEqual(f.writes, []);
});

test("workflow composes draft handler before unchanged review-ready policy", async () => {
  const workflow = await readFile(
    new URL("../workflows/close-stale-prs.yml", import.meta.url),
    "utf8"
  );
  assert.match(workflow, /if: github\.repository == 'radius-project\/radius'/);
  assert.match(workflow, /cancel-in-progress: false/);
  assert.match(workflow, /contents: read/);
  assert.match(workflow, /pull-requests: write/);
  assert.match(workflow, /persist-credentials: false/);
  assert.ok(
    workflow.indexOf("await closeStaleDrafts({ github, context, core })") <
      workflow.indexOf("- uses: actions/stale@")
  );
  assert.match(workflow, /exempt-draft-pr: true/);
  assert.match(workflow, /exempt-pr-labels: stale-draft/);
  assert.match(workflow, /days-before-pr-stale: 28/);
  assert.match(workflow, /days-before-pr-close: 1/);
  assert.match(workflow, /days-before-issue-stale: -1/);
  assert.match(workflow, /days-before-issue-close: -1/);
  assert.match(
    workflow,
    /- uses: actions\/stale@[^\n]+\n\s+if: \$\{\{ !cancelled\(\) \}\}/
  );
  const script = workflow.match(/script: \|\n((?: {12}.+\n)+)/)[1];
  const f = fixture();
  const previous = process.env.GITHUB_WORKSPACE;
  process.env.GITHUB_WORKSPACE = fileURLToPath(
    new URL("../../", import.meta.url)
  );
  try {
    const execute = new (Object.getPrototypeOf(
      async function () {}
    ).constructor)("github", "context", "core", script);
    await execute(
      f.github,
      { repo: { owner: "owner", repo: "repo" } },
      { info() {}, warning() {} }
    );
    assert.deepEqual(f.writes, ["label", "warn"]);
  } finally {
    if (previous === undefined) delete process.env.GITHUB_WORKSPACE;
    else process.env.GITHUB_WORKSPACE = previous;
  }
});
