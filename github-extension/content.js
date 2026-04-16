(function () {
  'use strict';

  var SKILL_URL =
    'https://raw.githubusercontent.com/kachawla/radius/gh-demo/.github/skills/app-modeling/SKILL.md';

  function getRepoInfo() {
    var match = window.location.pathname.match(/^\/([^/]+)\/([^/]+)/);
    if (!match) return null;
    var owner = match[1];
    var repo = match[2].replace(/\.git$/, '');
    var reserved = [
      'settings', 'notifications', 'sponsors', 'organizations',
      'orgs', 'topics', 'collections', 'explore', 'marketplace',
      'features', 'copilot',
    ];
    if (reserved.includes(owner)) return null;
    return { owner: owner, repo: repo };
  }

  function isRepoPage() {
    var path = window.location.pathname.split('/').filter(Boolean);
    if (path.length < 2) return false;
    if (path.length === 2) return true;
    if (path[2] === 'tree' || path[2] === 'blob') return true;
    return false;
  }

  function buildCopilotUrl(owner, repo) {
    var prompt = 'Create an application definition.\n\nRead ' + SKILL_URL;
    return 'https://github.com/copilot?repo=' +
      encodeURIComponent(owner + '/' + repo) +
      '&prompt=' + encodeURIComponent(prompt);
  }

  function createDeployButton(repoInfo) {
    if (document.getElementById('radius-deploy-container')) return null;

    var container = document.createElement('div');
    container.id = 'radius-deploy-container';
    container.className = 'radius-deploy-container';

    var btn = document.createElement('button');
    btn.className = 'radius-deploy-btn';
    btn.innerHTML = '&#9650; Deploy';
    btn.addEventListener('click', function (e) {
      e.stopPropagation();
      dropdown.classList.toggle('radius-dropdown-visible');
    });

    var dropdown = document.createElement('div');
    dropdown.className = 'radius-dropdown';

    var header = document.createElement('div');
    header.className = 'radius-dropdown-header';
    header.textContent = 'Deploy with Radius';
    dropdown.appendChild(header);

    var divider1 = document.createElement('div');
    divider1.className = 'radius-dropdown-divider';
    dropdown.appendChild(divider1);

    var appSection = document.createElement('div');
    appSection.className = 'radius-dropdown-section';
    appSection.textContent = 'Application';
    dropdown.appendChild(appSection);

    var defineApp = document.createElement('a');
    defineApp.className = 'radius-dropdown-item';
    defineApp.href = buildCopilotUrl(repoInfo.owner, repoInfo.repo);
    defineApp.textContent = 'Define an Application';
    defineApp.addEventListener('click', function () {
      dropdown.classList.remove('radius-dropdown-visible');
    });
    dropdown.appendChild(defineApp);

    var divider2 = document.createElement('div');
    divider2.className = 'radius-dropdown-divider';
    dropdown.appendChild(divider2);

    var envSection = document.createElement('div');
    envSection.className = 'radius-dropdown-section';
    envSection.textContent = 'Environments';
    dropdown.appendChild(envSection);

    var aws = document.createElement('div');
    aws.className = 'radius-dropdown-item radius-disabled';
    aws.textContent = 'Create AWS environment';
    dropdown.appendChild(aws);

    var azure = document.createElement('div');
    azure.className = 'radius-dropdown-item radius-disabled';
    azure.textContent = 'Create Azure environment';
    dropdown.appendChild(azure);

    var gcp = document.createElement('div');
    gcp.className = 'radius-dropdown-item radius-disabled';
    gcp.textContent = 'Create Google Cloud environment';
    dropdown.appendChild(gcp);

    container.appendChild(btn);
    container.appendChild(dropdown);

    document.addEventListener('click', function (e) {
      if (!container.contains(e.target)) {
        dropdown.classList.remove('radius-dropdown-visible');
      }
    });

    return container;
  }

  function inject() {
    if (!isRepoPage()) return;
    if (document.getElementById('radius-deploy-container')) return;

    var repoInfo = getRepoInfo();
    if (!repoInfo) return;

    var deployBtn = createDeployButton(repoInfo);
    if (!deployBtn) return;

    var actions =
      document.querySelector('.pagehead-actions') ||
      document.querySelector('[data-testid="repo-header-actions"]');
    if (actions) {
      var li = document.createElement('li');
      li.appendChild(deployBtn);
      actions.prepend(li);
      return;
    }

    var codeBtn = document.querySelector('[data-hotkey="t"]');
    if (codeBtn && codeBtn.parentElement) {
      codeBtn.parentElement.insertBefore(deployBtn, codeBtn);
      return;
    }

    var headerEl = document.querySelector('.repository-content') ||
                   document.querySelector('[class*="Layout-main"]') ||
                   document.querySelector('main');
    if (headerEl) {
      headerEl.insertBefore(deployBtn, headerEl.firstChild);
    }
  }

  inject();

  var observer = new MutationObserver(function () {
    if (isRepoPage() && !document.getElementById('radius-deploy-container')) {
      inject();
    }
  });
  observer.observe(document.body, { childList: true, subtree: true });
})();