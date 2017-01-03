'use strict';

(function(exports){

  exports.create = function(server){

    function handleResp(status, resp, resolve, reject, cb) {
      if (status === 200) {
        resolve(resp);
      } else {
        reject(resp);
      }

      if (cb) {
        cb(status, resp);
      }
    }

    function getOAuthClientID(provider, cb) {
      return new Promise(function(resolve, reject) {
        var xhr = new XMLHttpRequest();
        xhr.onload = function() {
          var resp = JSON.parse(xhr.responseText);
          handleResp(xhr.status, resp, resolve, reject, cb);
        }
        xhr.onerror = function(e) {
          reject({status: e.target.status});
        }
        xhr.open(
          "GET",
          "http://" + server + "/api/auth/" + provider + "/client_id",
          true
        );
        xhr.send(null);
      });
    }

    function authenticate(provider, redirectURI, code, cb) {
      return new Promise(function(resolve, reject) {
        var xhr = new XMLHttpRequest();
        xhr.onload = function() {
          var resp = JSON.parse(xhr.responseText);
          handleResp(xhr.status, resp, resolve, reject, cb);
        }
        xhr.onerror = function(e) {
          reject({status: e.target.status});
        }
        var url = URI("http://" + server + "/api/auth/" + provider + "/authenticate")
          .addSearch("code", code)
          .addSearch("redirect_uri", chrome.identity.getRedirectURL())
          .toString();
        xhr.open(
          "POST",
          url,
          true
        );
        xhr.send(null);
      });
    }

    function createGMClientLoggedIn() {
      return new Promise(function(resolve, reject) {
        var tokenKey = 'token_' + server;
        chrome.storage.sync.get(tokenKey, function(v) {
          if (tokenKey in v) {
            // We have a token
            resolve(_createGMClientLoggedIn());
          } else {
            // We don't have a token, just return null
            // TODO: so far, we anyway return a valid instance with hardcoded creds
            //resolve(null);
            resolve(_createGMClientLoggedIn());
          }
        });
      });
    }

    function login(provider) {
      return new Promise(function(resolve, reject) {
        getOAuthClientID(provider).then(function(resp) {
          var url = URI("https://accounts.google.com/o/oauth2/auth")
            .addSearch("scope", "email")
            .addSearch("redirect_uri", chrome.identity.getRedirectURL())
            .addSearch("client_id", resp.clientID)
            .addSearch("response_type", "code")
            .toString();
          console.log(url);
          chrome.identity.launchWebAuthFlow(
            {
              url: url,
              interactive: true
            },
            function(responseURL) {
              var uri = new URI(responseURL);
              var queryParams = uri.search(true);

              console.log(queryParams.code);

              authenticate("google", responseURL, queryParams.code).then(function(resp) {
                // Got a token
                console.log('token!', resp);
                resolve();
              }).catch(function(e) {
                reject(e);
              });
            }
          );
        }).catch(function(e) {
          reject(e);
        });
      });
    }

    function _createGMClientLoggedIn() {
      // TODO: use real token
      var user = "alice"
      var password = "alice"

      var msgID = 0;
      var pendingRequests = {};
      var isConnected = false;
      var onConnectedCB = undefined;
      var artificialDelay = 0;
      var ws = new WebSocket(
        "ws://" + user + ":" + password + "@" + server + "/api/my/wsconnect"
      );

      if (server.substring(0, 9) === "localhost") {
        artificialDelay = 150;
      }

      ws.onopen = function() {
        console.log("Connection opened");
        isConnected = true;
        if (onConnectedCB) {
          onConnectedCB();
        }
      };
      ws.onclose = function() {
        console.log("Connection closed");
        isConnected = false;
        // TODO: reconnect
      };

      ws.onmessage = function(evt) {
        console.log("ws event:", evt);
        var msg = JSON.parse(evt.data);
        var sID = String(msg.id);
        if (sID in pendingRequests) {
          var pr = pendingRequests[sID];
          if (typeof(pr) === 'object') {
            pr.cb(msg.status, msg.body);
          }
          delete pendingRequests[sID];
        }
      };

      function send(msg, cb) {
        msgID++;
        var sID = String(msgID)
        if (sID in pendingRequests) {
          throw Error("should never happen");
        }

        msg.id = msgID;

        pendingRequests[sID] = {
          cb: cb,
        };

        var f = function() {
          ws.send(JSON.stringify(msg));
        };

        if (artificialDelay == 0) {
          f();
        } else {
          setTimeout(f, artificialDelay);
        }
      }

      function onConnected(invokeIfAlreadyConnected, cb) {
        onConnectedCB = cb;
        if (onConnectedCB && isConnected && invokeIfAlreadyConnected) {
          onConnectedCB();
        }
      }

      function getTagsTree(cb) {
        // TODO: support subpath, support no-subtags
        console.log("getTagsTree is called");
        send({
          path: "/tags",
          method: "GET",
        }, cb);
      }

      function getTag(tagPathOrID, cb) {
        // TODO: support subpath, support no-subtags
        console.log("getTag is called");
        send({
          path: ["/tags", tagPathOrID].join("/"),
          method: "GET",
          shape: "single",
        }, cb);
      }

      function getTagsByPattern(pattern, allowNew, cb) {
        console.log("getTagsByPattern is called, pattern:", pattern);
        var values = {
          shape: "flat",
          pattern: pattern,
        };
        if (allowNew) {
          values.allow_new = "1";
        }
        send({
          path: "/tags",
          method: "GET",
          values: values,
        }, cb);
      }

      function addTag(parentTagPath, data, cb) {
        console.log("addTag is called:", parentTagPath, data);
        send({
          path: ["/tags", parentTagPath].join("/"),
          method: "POST",
          body: {
            names: data.names,
            description: data.description,
            createIntermediary: data.createIntermediary,
          },
        }, cb);
      }

      function updateTag(tagPathOrID, tagData, cb) {
        console.log("updateTag is called, tagPathOrID:", tagPathOrID, ", tagData:", tagData);
        send({
          path: ["/tags", tagPathOrID].join("/"),
          body: {
            names: tagData.names,
            description: tagData.description,
            parentTagID: Number(tagData.parentTagID),
          },
          method: "PUT"
        }, cb);
      }

      function getTaggedBookmarks(tagIDs, cb) {
        console.log("getTaggedBookmarks is called, tagIDs:", tagIDs);
        send({
          path: "/bookmarks",
          method: "GET",
          values: {
            tag_id: tagIDs,
          }
        }, cb);
      }

      function getBookmarksByURL(url, cb) {
        console.log("getBookmarksByURL is called, url:", url);
        send({
          path: "/bookmarks",
          method: "GET",
          values: {
            url: [url],
          }
        }, cb);
      }

      function getBookmarkByID(bookmarkID, cb) {
        console.log("getBookmarkByID is called, bookmarkID:", bookmarkID);
        send({
          path: "/bookmarks/" + bookmarkID,
          method: "GET"
        }, cb);
      }

      function addBookmark(bkmData, cb) {
        console.log("addBookmark is called:", ", bkmData:", bkmData);
        send({
          path: "/bookmarks",
          body: {
            url: bkmData.url,
            title: bkmData.title,
            comment: bkmData.comment,
            tagIDs: bkmData.tagIDs,
          },
          method: "POST"
        }, cb);
      }

      function updateBookmark(bookmarkID, bkmData, cb) {
        console.log("updateBookmark is called, bookmarkID:", bookmarkID, ", bkmData:", bkmData);
        send({
          path: "/bookmarks/" + bookmarkID,
          body: {
            url: bkmData.url,
            title: bkmData.title,
            comment: bkmData.comment,
            tagIDs: bkmData.tagIDs,
          },
          method: "PUT"
        }, cb);
      }

      return {
        onConnected: onConnected,
        getTagsTree, getTagsTree,
        getTagsByPattern: getTagsByPattern,
        getTag: getTag,
        addTag: addTag,
        updateTag, updateTag,
        getTaggedBookmarks: getTaggedBookmarks,
        getBookmarksByURL: getBookmarksByURL,
        getBookmarkByID: getBookmarkByID,
        addBookmark: addBookmark,
        updateBookmark: updateBookmark,
      };
    }

    return {
      createGMClientLoggedIn: createGMClientLoggedIn,
      login: login,
      getOAuthClientID: getOAuthClientID,
    };

  };

})(typeof exports === 'undefined' ? this['gmClient']={} : exports);
