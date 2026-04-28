(function () {
  'use strict';

  var tokenInput = document.getElementById('token');
  var saveBtn = document.getElementById('save');
  var statusEl = document.getElementById('status');

  chrome.storage.sync.get(['githubToken'], function (result) {
    if (result.githubToken) {
      tokenInput.value = result.githubToken;
    }
  });

  saveBtn.addEventListener('click', function () {
    var token = tokenInput.value.trim();
    if (!token) return;
    chrome.storage.sync.set({ githubToken: token }, function () {
      statusEl.style.display = 'block';
      setTimeout(function () {
        statusEl.style.display = 'none';
      }, 2000);
    });
  });
})();
