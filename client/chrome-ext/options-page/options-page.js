(function() {
  // Saves options to chrome.storage
  function saveOptions() {
    var server = undefined;
    var serverSSL = false;

    var serverURL = document.getElementById('server_addr').value;

    // TODO: proper URL parsing
    if (serverURL.slice(0, 8) === "https://") {
      server = serverURL.slice(8);
      serverSSL = true;
    } else if (serverURL.slice(0, 7) === "http://") {
      server = serverURL.slice(7);
      serverSSL = false;
    } else {
      server = serverURL;
      serverSSL = false;
    }

    gmOptions.setOptions({
      server: server,
      serverSSL: serverSSL,
    }, function() {
      // Update status to let user know options were saved.
      var status = document.getElementById('status');
      status.textContent = 'Options saved.';
      setTimeout(function() {
        status.textContent = '';
      }, 750);
    })
  }

  // Restores select box and checkbox state using the preferences
  // stored in chrome.storage.
  function restoreOptions() {
    // Use default value color = 'red' and likesColor = true.
    gmOptions.getOptions(function(opts) {
      var serverURL = (opts.serverSSL ? "https" : "http") + "://" + opts.server
      document.getElementById('server_addr').value = serverURL;
    });
  }

  document.addEventListener('DOMContentLoaded', restoreOptions);
  document.getElementById('save').addEventListener('click', saveOptions);
})();
