(function () {
  'use strict';

  var SKILL_URL =
    'https://raw.githubusercontent.com/kachawla/radius/demo/.github/skills/app-modeling/SKILL.md';

  var POLL_INTERVAL_MIN = 0.25; // ~15 seconds
  var MAX_POLL_TIME = 600000; // 10 minutes

  function apiHeaders(token) {
    return {
      'Accept': 'application/vnd.github+json',
      'Authorization': 'Bearer ' + token,
      'X-GitHub-Api-Version': '2022-11-28',
      'Content-Type': 'application/json',
    };
  }

  function sendStatus(tabId, status, data) {
    chrome.tabs.sendMessage(tabId, { action: 'copilotStatus', status: status, data: data }).catch(function () {
      console.log('[Radius] Could not send status to tab ' + tabId);
    });
  }

  function startPolling(token, owner, repo, issueNumber, tabId) {
    var state = {
      token: token,
      owner: owner,
      repo: repo,
      issueNumber: issueNumber,
      tabId: tabId,
      startTime: Date.now(),
    };
    chrome.storage.local.set({ radiusPollState: state }, function () {
      console.log('[Radius] Polling started for issue #' + issueNumber);
      doPoll();
      chrome.alarms.create('radiusPoll', { periodInMinutes: POLL_INTERVAL_MIN });
    });
  }

  function stopPolling() {
    chrome.alarms.clear('radiusPoll');
    chrome.storage.local.remove('radiusPollState');
    console.log('[Radius] Polling stopped');
  }

  function doPoll() {
    chrome.storage.local.get(['radiusPollState'], function (result) {
      var state = result.radiusPollState;
      if (!state) {
        chrome.alarms.clear('radiusPoll');
        return;
      }

      var token = state.token;
      var owner = state.owner;
      var repo = state.repo;
      var issueNumber = state.issueNumber;
      var tabId = state.tabId;

      if (Date.now() - state.startTime > MAX_POLL_TIME) {
        sendStatus(tabId, 'timeout', {
          message: 'Copilot is still working. Check the issue for progress.',
          issueUrl: 'https://github.com/' + owner + '/' + repo + '/issues/' + issueNumber,
        });
        stopPolling();
        return;
      }

      // Scan open PRs for the most recent one by Copilot
      fetch('https://api.github.com/repos/' + owner + '/' + repo + '/pulls?state=open&per_page=10&sort=created&direction=desc', {
        headers: apiHeaders(token),
      })
        .then(function (r) { return r.json(); })
        .then(function (pulls) {
          console.log('[Radius] Polling PRs, found:', pulls.length);
          var match = null;
          for (var i = 0; i < pulls.length; i++) {
            var pr = pulls[i];
            var login = pr.user && pr.user.login;
            var byBot = login === 'copilot-swe-agent[bot]' ||
                        login === 'Copilot' ||
                        login === 'copilot';

            console.log('[Radius] PR #' + pr.number + ' by ' + login + ' draft=' + pr.draft);

            if (byBot) {
              match = pr;
              console.log('[Radius] Matched PR #' + pr.number);
              break;
            }
          }

          if (!match) {
            console.log('[Radius] No Copilot PR found yet');
            return; // alarm will retry
          }

          // Fetch full PR details
          return fetch('https://api.github.com/repos/' + owner + '/' + repo + '/pulls/' + match.number, {
            headers: apiHeaders(token),
          })
            .then(function (r) { return r.json(); })
            .then(function (fullPR) {
              console.log('[Radius] Full PR #' + fullPR.number + ' draft=' + fullPR.draft);
              var hasReviewers = fullPR.requested_reviewers && fullPR.requested_reviewers.length > 0;

              if (fullPR.draft && !hasReviewers) {
                sendStatus(tabId, 'pr_found', {
                  message: 'Copilot opened a draft PR, waiting for it to finish...',
                  prUrl: fullPR.html_url,
                });
                return; // alarm will retry
              }

              // Copilot is done — merge
              stopPolling();
              if (fullPR.draft) {
                markReadyAndMerge(token, owner, repo, fullPR.node_id, fullPR.number, tabId);
              } else {
                mergePR(token, owner, repo, fullPR.number, tabId);
              }
            });
        })
        .catch(function (err) {
          console.log('[Radius] Poll error:', err.message);
        });
    });
  }

  chrome.alarms.onAlarm.addListener(function (alarm) {
    if (alarm.name === 'radiusPoll') {
      doPoll();
    }
  });

  function markReadyAndMerge(token, owner, repo, prNodeId, prNumber, tabId) {
    sendStatus(tabId, 'pr_found', {
      message: 'Copilot finished — marking PR ready and merging...',
      prUrl: 'https://github.com/' + owner + '/' + repo + '/pull/' + prNumber,
    });

    fetch('https://api.github.com/graphql', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer ' + token,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query: 'mutation($id: ID!) { markPullRequestReadyForReview(input: { pullRequestId: $id }) { pullRequest { isDraft } } }',
        variables: { id: prNodeId },
      }),
    })
      .then(function (r) { return r.json(); })
      .then(function (result) {
        if (result.errors && result.errors.length > 0) {
          throw new Error(result.errors[0].message);
        }
        mergePR(token, owner, repo, prNumber, tabId);
      })
      .catch(function (err) {
        sendStatus(tabId, 'merge_error', {
          message: 'Could not mark PR ready: ' + err.message,
          prUrl: 'https://github.com/' + owner + '/' + repo + '/pull/' + prNumber,
        });
      });
  }

  function mergePR(token, owner, repo, prNumber, tabId) {
    fetch('https://api.github.com/repos/' + owner + '/' + repo + '/pulls/' + prNumber + '/merge', {
      method: 'PUT',
      headers: apiHeaders(token),
      body: JSON.stringify({
        commit_title: 'Add Radius application definition',
        merge_method: 'squash',
      }),
    })
      .then(function (response) {
        if (!response.ok) {
          return response.json().then(function (err) {
            throw new Error(err.message || 'Merge failed');
          });
        }
        return response.json();
      })
      .then(function () {
        sendStatus(tabId, 'merged', {
          message: 'Application definition merged!',
          fileUrl: 'https://github.com/' + owner + '/' + repo + '/applications',
        });
      })
      .catch(function (err) {
        sendStatus(tabId, 'merge_error', {
          message: 'PR ready but auto-merge failed: ' + err.message,
          prUrl: 'https://github.com/' + owner + '/' + repo + '/pull/' + prNumber,
        });
      });
  }

  chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
    if (message.action !== 'createCopilotTask') return false;

    var owner = message.owner;
    var repo = message.repo;
    var tabId = sender.tab && sender.tab.id;

    chrome.storage.sync.get(['githubToken'], function (result) {
      var token = result.githubToken;
      if (!token) {
        sendResponse({ error: 'no_token', message: 'GitHub token not configured. Click the Radius extension icon to set it up.' });
        return;
      }

      var issueTitle = 'Create Radius application definition';
      var issueBody =
        'Create an application definition for this repository.\n\n' +
        'Read the skill instructions at: ' + SKILL_URL + '\n\n' +
        'Follow the skill instructions to analyze this repository and generate a `.radius/app.bicep` file.\n' +
        'Create a pull request with the generated file.';

      fetch('https://api.github.com/repos/' + owner + '/' + repo + '/issues', {
        method: 'POST',
        headers: apiHeaders(token),
        body: JSON.stringify({
          title: issueTitle,
          body: issueBody,
          assignees: ['copilot-swe-agent[bot]'],
          agent_assignment: {
            target_repo: owner + '/' + repo,
            base_branch: 'main',
            custom_instructions: 'Read the skill at ' + SKILL_URL + ' and follow its instructions exactly.',
            custom_agent: '',
            model: '',
          },
        }),
      })
        .then(function (response) {
          if (!response.ok) {
            return response.json().then(function (err) {
              throw new Error(err.message || 'API request failed with status ' + response.status);
            });
          }
          return response.json();
        })
        .then(function (issue) {
          sendResponse({
            success: true,
            issueNumber: issue.number,
            issueUrl: issue.html_url,
          });
          startPolling(token, owner, repo, issue.number, tabId);
        })
        .catch(function (err) {
          sendResponse({ error: 'api_error', message: err.message });
        });
    });

    return true;
  });
})();
