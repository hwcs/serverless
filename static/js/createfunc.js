require (['require-config'], function () {

require(['jquery', 'codemirror', 'js/common', 'js/left-nav', 'js/top-nav', 'codemirror/mode/python/python', "codemirror/mode/javascript/javascript", 'bootstrap'], function($, CodeMirror, common, leftnav, topnav) {

$(document).ready(function () {
  // Load nav bar
  topnav.loadTopNav();
  leftnav.loadLeftNav("left-nav-func");
  $("#fun-name-nav").html("Create");
  changeCodeContent ();

  var editor = CodeMirror.fromTextArea(document.getElementById("func-code"), {
    mode: {name: "python",
           version: 3,
           singleLineStringErrors: false},
    lineNumbers: true,
    indentUnit: 4,
    matchBrackets: true
  });

  $.ajax({
      type: 'GET',
      url: "/data/function.py",
      dataType: "text",
      success: function(resultData) {
        editor.getDoc().setValue(resultData);
        editor.refresh();
      },
      failure: function(errMsg) {
            alert(errMsg);
      },
      error: function(e) {
        alert(e);
      }
    });

    $("#btn-ok").click (function () {
      var funcname = $("#func-meta-name").val();
      var codeType = $("#func-code-type").val ();
      var memorySize = parseInt($("#func-meta-mem").val());
      var timeout = parseInt($("#func-meta-timeout").val());

      postData = 	{
    	  "Description": $("#func-meta-desc").val(),
          "FunctionName": funcname,
          "Handler": $("#func-meta-handler").val(),
          "MemorySize": memorySize,
          "Timeout": timeout,
          "Runtime": $("#func-meta-runtime").val(),

    		};

      if (codeType == "inline") {
        //postData["File"] = editor.getDoc().getValue ();
        postData["FuncCode"] = {
            "CodeType": codeType,
            "File" : editor.getDoc().getValue (),
        };
        $.ajax({
      		type: "POST",
      		url: "/functions/",
      		contentType: "application/json",
      		processData: false,
      		data: JSON.stringify(postData),

            statusCode: {
                400: function() {
                    alert( "Code size exceed max size 100M ");
                },
                409: function() {
                    alert( "Function existed: " + funcname );
                },
                500: function() {
                    alert( "Internal Server Error" );
                },
                200: function() {
                    alert( "Save succeed, functiona name: " + funcname);
                    window.location.href = "functions.html";
                }
            },
 /*
  		      success: function(resultData) {
                  funcName = resultData.FunctionName;
  		        alert("Save succeed, Functiona Name: " + funcName);
  		      },
  		      failure: function(errMsg) {
                  alert("failure")
  		            alert(errMsg);
  		      },
  		      error: function(e) {
                  alert("error")
  		        alert(e);
            }
            */
  		 });
      } else if (codeType == "upload") {
        var fileInput = document.getElementById('func-code-file');

        if (fileInput.files.length == 0) {
          alert ("Please select function package zip file!");
        } else {
          var file = fileInput.files[0];

          var reader = new FileReader();

          reader.onerror = function errorHandler(evt) {
            switch(evt.target.error.code) {
              case evt.target.error.NOT_FOUND_ERR:
                alert('File Not Found!');
                break;
              case evt.target.error.NOT_READABLE_ERR:
                alert('File is not readable');
                break;
              case evt.target.error.ABORT_ERR:
                break;
              default:
                alert('An error occurred reading this file.');
            };
          }

          reader.onload = function (result) {
              if (reader.readyState == FileReader.DONE) {
                //postData["File"] = arrayBufferToBase64 (result.target.result);
                postData["FuncCode"] = {
                    "CodeType": "upload",
                    "File" : common.arrayBufferToBase64 (result.target.result),
                };
                $.ajax({
              		type: "POST",
              		url: "/functions/",
              		contentType: "application/json",
              		processData: false,
              		data: JSON.stringify(postData),
                    /*
          		      success: function(resultData) {
          		        alert("Save succeed.");
          		      },
          		      failure: function(errMsg) {
          		            alert(errMsg);
          		      },
          		      error: function(e) {
          		        alert(e);
                    }*/
                    statusCode: {
                        400: function() {
                            alert( "Code size exceed max size 100M ");
                        },
                        409: function() {
                            alert( "Function existed: " + funcname );
                        },
                        500: function() {
                            alert( "Internal Server Error" );
                        },
                        200: function() {
                            alert( "Save succeed, functiona name: " + funcname);
                            window.location.href = "functions.html";
                        }
                    },
          		 });
             } else {
               alert ("Read file " + fileName + " error!");
             }
           }
           reader.readAsArrayBuffer(file);
        }
      }
    });

    function changeCodeContent () {
      var codeType = $("#func-code-type").val ();
      if (codeType == "inline") {
        $("#func-code-online").show ();
        $("#func-code-upload").hide ();
      } else if (codeType == "upload") {
        $("#func-code-online").hide ();
        $("#func-code-upload").show ();
      }
    }

    $("#func-code-type").change (function () {
      changeCodeContent ();
    });

    $("#btn-cancel").click (function () {
      window.location.href = "index.html";
    });
});

});

});

