'use strict';

(function(exports){

  var contentElem = undefined;
  var gmClient = undefined;

  var rootTagKey = undefined;
  var keyToTag = {};

  function init(_gmClient, _contentElem, srcDir, queryParams, curTabData) {
    contentElem = _contentElem;
    gmClient = _gmClient;
    var tagsTreeDiv = contentElem.find('#tags_tree_div')

    gmClient.getTagsTree(function(status, resp) {
      if (status == 200) {
        var treeData = convertTreeData(resp, true);

        tagsTreeDiv.fancytree({
          extensions: ["edit", "table", "dnd"],
          edit: {
            adjustWidthOfs: 4,   // null: don't adjust input size to content
            inputCss: { minWidth: "3em" },
            triggerStart: ["f2", "shift+click", "mac+enter"],
            beforeEdit: function(event, data) {
              if (data.node.key === rootTagKey) {
                return false;
              }
              return true;
            },
            edit: $.noop,        // Editor was opened (available as data.input)
            beforeClose: $.noop, // Return false to prevent cancel/save (data.input is available)
            save: saveTag,       // Save data.input.val() or return false to keep editor open
            close: $.noop,       // Editor was removed
          },
          table: {
          },
          dnd: {
            // Available options with their default:
            autoExpandMS: 1000,   // Expand nodes after n milliseconds of hovering
            draggable: null,      // Additional options passed to jQuery UI draggable
            droppable: null,      // Additional options passed to jQuery UI droppable
            focusOnClick: false,  // Focus, although draggable cancels mousedown event (#270)
            preventRecursiveMoves: true, // Prevent dropping nodes on own descendants
            preventVoidMoves: true,      // Prevent dropping nodes 'before self', etc.
            smartRevert: true,    // set draggable.revert = true if drop was rejected

            // Events that make tree nodes draggable
            dragStart: function(node, data) {
              return true;
            },
            dragStop: null,       // Callback(sourceNode, data)
            initHelper: null,     // Callback(sourceNode, data)
            updateHelper: null,   // Callback(sourceNode, data)

            // Events that make tree nodes accept draggables
            dragEnter: function(node, data) {
              // allow only moving nodes under other nodes; do not allow
              // reordering.
              // (to allow reordering, the returned array should also contain
              // "before", "after".)
              return ["over"];
            },
            dragExpand: null,     // Callback(targetNode, data), return false to prevent autoExpand
            dragOver: null,       // Callback(targetNode, data)
            dragDrop: function(node, data) {
              // This function MUST be defined to enable dropping of items on the tree.
              // data.hitMode is 'before', 'after', or 'over'.
              // We could for example move the source to the new target:
              var oldParent = data.otherNode.parent;
              var subj = data.otherNode;

              if (confirm('move "' + subj.title + '" under "' + data.node.title + '"?')) {
                subj.moveTo(node, data.hitMode);
                subj.makeVisible();

                gmClient.updateTag(subj.key, {
                  parentTagID: data.node.key,
                }, function(status, resp) {
                  if (status == 200) {
                    // move succeeded, do nothing here
                  } else {
                    // TODO: show error
                    alert(JSON.stringify(resp));
                    subj.moveTo(oldParent, "over");
                  }

                  $(data.node.span).removeClass("pending");
                });

                // Here is the code to move the node back, if actual move
                // fails on the server:
                /*
                setTimeout(function() {
                  // NOTE: there is an issue with `moveTo()`: it moves the node
                  // to nowhere if the node is not visible. So, first of all
                  // we need to check if node is not visible, and if so,
                  // ensure it is visible:
                  var nodeToClose = undefined;
                  if (!subj.isVisible()) {
                    subj.makeVisible();
                    nodeToClose = subj.parent;
                  }

                  // Now, move node
                  subj.moveTo(oldParent, "over");

                  // And now, if it wasn't visible, close the parent back.
                  if (nodeToClose !== undefined) {
                    nodeToClose.setExpanded(false);
                  }
                }, 1000);
                */
              }
            },
            dragLeave: null       // Callback(targetNode, data)
          },
          source: {
            children: [treeData],
          },
          renderColumns: function(event, data) {
            if (data.node.key !== rootTagKey) {
              var node = data.node;
              var $tdList = $(node.tr).find(">td");
              var $ctrlCol = $tdList.eq(2);
              $ctrlCol.text("");
              $("<a/>", {
                href: "#",
                text: "[edit]",
                click: function() {
                  gmPageWrapper.openPageEditTag(data.node.key);
                },
              }).appendTo($ctrlCol);
            }
          },
        });

        var tree = tagsTreeDiv.fancytree("getTree");
        var rootNode = tree.getNodeByKey(rootTagKey);
        rootNode.setExpanded(true);
      } else {
        // TODO: show error
        alert(JSON.stringify(resp));
      }
    })
  }

  function convertTreeData(tagsTree, isRoot) {
    var key = String(tagsTree.id);

    var ret = {
      title: tagsTree.names.join(", "),
      key: key,
    };

    keyToTag[key] = tagsTree;

    if (isRoot) {
      ret.title = "my tags";
      rootTagKey = key;
    }
    if ("subtags" in tagsTree) {
      ret.children = tagsTree.subtags.map(function(a) {
        return convertTreeData(a, false);
      });
      ret.folder = true;
    }
    return ret;
  }

  // see https://github.com/mar10/fancytree/wiki/ExtEdit for argument details
  function saveTag(event, data) {
    $(data.node.span).addClass("pending");
    var val = data.input.val();
    var prevVal = data.orgTitle;
    //console.log('saveTag event', event)
    //console.log('saveTag data', data)

    gmClient.updateTag(String(data.node.key), {
      names: val.split(",").map(function(a) {
        return a.trim();
      }),
    }, function(status, resp) {
      if (status == 200) {
        // update succeeded, do nothing here
      } else {
        // TODO: show error
        alert(JSON.stringify(resp));
        data.node.setTitle(prevVal);
      }

      $(data.node.span).removeClass("pending");
    });

    // Optimistically assume that save will succeed. Accept the user input
    return true;
  }

  exports.init = init;

})(typeof exports === 'undefined' ? this['gmTagsTree']={} : exports);
