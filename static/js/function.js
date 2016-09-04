require (['require-config'], function () {

require(['jquery', 'codemirror', 'js/common', 'js/left-nav', 'js/top-nav', 'codemirror/mode/python/python', "codemirror/mode/javascript/javascript", 'bootstrap'], function($, CodeMirror, common, leftnav, topnav) {

$(document).ready(function () {
  var testFunction = function () {
    var payload = window.localStorage.getItem("test-event/" + funcName);
    alert(payload);
    var start = new Date().getTime();
    $.ajax({
      type: "POST",
      url: "/functions/" + funcName + "/invocations",
      contentType: "application/json",
      resultData: "json",
      data: payload,
        success: function(resultData) {
          $("#div-func-exec-result").show ();
          $("#func-exec-status").val("succeed");
          $("#func-exec-time").val("" + new Date().getTime() - start + " ms");
          $("#func-exec-result").val(JSON.stringify(resultData));
        },
        error: function(jqXHR, textStatus, errorThrown) {
          $("#div-func-exec-result").show ();
          $("#func-exec-status").val("error");
          $("#func-exec-time").val("" + new Date().getTime() - start + " ms");
          $("#func-exec-result").val("responseText : " + jqXHR.responseText + "\ntextStatus : " + textStatus + "\nerrorThrown : " + errorThrown);
        }
   });
  };

  var loadTestEvent = function (parentId, func) {
  	var funcName = func;

      var editor = CodeMirror.fromTextArea(document.getElementById("func-test-event-editor"), {
  		mode: "application/ld+json",
  		lineWrapping: true,
  		autoCloseBrackets: true,
  		lineNumbers: true,
  		matchBrackets: true
  	});

  	var oldContent = window.localStorage.getItem("test-event/" + funcName);
  	if (oldContent != null) {
  		editor.getDoc().setValue (oldContent);
  		editor.refresh ();
  	}

  	$("#btn-test-event").click (function () {
  		$("#func-test-event").modal ();
  	});

  	$("#func-test-event").on('shown.bs.modal', function () {
  		editor.refresh ();
  	});

    function resetEventModal () {
      if (eventModal == "open") {
        url = "/function.html?name=" + funcName;
        window.history.replaceState(null,null,url);
      }
    };

  	$("#btn-test-save").click (function () {
  		window.localStorage.setItem("test-event/" + funcName, editor.getDoc().getValue ());
  		$("#func-test-event").modal('hide');
      resetEventModal ();
  	});

  	$("#btn-test-save-test").click (function () {
  		window.localStorage.setItem("test-event/" + funcName, editor.getDoc().getValue ());
  		$("#func-test-event").modal('hide');
      resetEventModal ();
      testFunction ();
  	});
  };

  // Load nav bar
  topnav.loadTopNav();
  leftnav.loadLeftNav("left-nav-func");
  funcName=common.getUrlVar('name');
  eventModal=common.getUrlVar('eventModal');
  $("#fun-name-nav").html(funcName);
  loadTestEvent("test-evt-dlg", funcName);
  $("#div-func-exec-result").hide ();

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

  // Load func configuration
  //funcConfUrl = "data/function-" + funcName + ".json";
  funcConfUrl = "/functions/" + funcName;
  $.ajax({
      type: 'GET',
      url: funcConfUrl,
      dataType: "json",
      success: function(resultData) {
        // To do set func configuration to call
        $("#func-meta-runtime").val(resultData.Runtime);
        $("#func-meta-handler").val(resultData.Handler);
        $("#func-meta-desc").val(resultData.Description);
        $("#func-meta-mem").val(resultData.MemorySize);
        $("#func-meta-timeout").val(resultData.Timeout);
        $("#http-trigger-url").html("https://ec2-52-25-244-252.us-west-2.compute.amazonaws.com:8080/runLambda/" + funcName);
      },
      failure: function(errMsg) {
            alert(errMsg);
      },
      error: function(e) {
        alert(e);
      }
    });

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
      url: "/functions/" + funcName + "/code" ,
      dataType: "text" ,
      success: function(resultData) {
        var data = JSON.parse(resultData)
        codetype = data["CodeType"];

        $("#func-code-type").val(codetype);
        if(codetype == "inline")
        {
            //editor.getDoc().setValue(resultData);
            editor.getDoc().setValue(data["Code"]);
            editor.refresh();
        }

        changeCodeContent ();
      },
      failure: function(errMsg) {
            alert(errMsg);
      },
      error: function(e) {
        alert(e);
      }
    });

    if (eventModal == "open") {
      $("#btn-test-event").click ();
    }

    var saveFunction = function () {
      var codeType = $("#func-code-type").val ();
      var memorySize = parseInt($("#func-meta-mem").val());
      var timeout = parseInt($("#func-meta-timeout").val());

      postData = 	{
    	  "Description": $("#func-meta-desc").val(),
          "Handler": $("#func-meta-handler").val(),
          "MemorySize": memorySize,
          "Timeout": timeout,
          "Runtime": $("#func-meta-runtime").val(),
    		};

      // Update function configuration
    	$.ajax({
    		type: "PUT",
    		url: "/functions/" + funcName + "/config",
    		contentType: "application/json",
    		processData: false,
    		data: JSON.stringify(postData),
		      success: function(resultData) {
		      },
          error: function(jqXHR, textStatus, errorThrown) {
            alert ("responseText : " + jqXHR.responseText + "\ntextStatus : " + textStatus + "\nerrorThrown : " + errorThrown);
          }
		 });

     // update function code
     postData = 	{
         "FunctionName": funcName,
         "Runtime": $("#func-meta-runtime").val (),
       };

     if (codeType == "inline") {
         postData["FuncCode"] = {
             "CodeType": codeType,
             "File" : editor.getDoc().getValue (),
         };
       $.ajax({
         type: "PUT",
         url: "/functions/" + funcName + "/code",
         contentType: "application/json",
         processData: false,
         data: JSON.stringify(postData),
           success: function(resultData) {
               alert("Save function : " + funcName + ", succeed");
           },
           error: function(jqXHR, textStatus, errorThrown) {
             alert ("responseText : " + jqXHR.responseText + "\ntextStatus : " + textStatus + "\nerrorThrown : " + errorThrown);
           }
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
                 postData["FuncCode"] = {
                     "CodeType": codeType,
                     "File" : common.arrayBufferToBase64 (result.target.result),
                 };

               $.ajax({
                 type: "PUT",
                 url: "/functions/" + funcName + "/code",
                 contentType: "application/json",
                 processData: false,
                 data: JSON.stringify(postData),
                   success: function(resultData) {
                   },
                   error: function(jqXHR, textStatus, errorThrown) {
                     alert ("responseText : " + jqXHR.responseText + "\ntextStatus : " + textStatus + "\nerrorThrown : " + errorThrown);
                   }
              });
            } else {
              alert ("Read file " + fileName + " error!");
            }
          }
          reader.readAsArrayBuffer(file);
       }
     }
   };

   $("#btn-func-save").click (function () {
     saveFunction ();
   });

   $("#btn-func-savetest").click (function () {
     saveFunction ();
     testFunction ();
   });

    $("#func-code-type").change (function () {
      changeCodeContent ();
    });

    $("#btn-func-test").click (function () {
      testFunction ();
    });

    $("#btn-del-func").click(function () {
        $.ajax({
            type: 'DELETE',
            url: "/functions/" + funcName,
            dataType: "json",
            success: function(resultData) {
              // delete function
              alert("Delete function : " + funcName);
              window.location.href = "functions.html";
            },
            error: function(jqXHR, textStatus, errorThrown) {
              alert ("responseText : " + jqXHR.responseText + "\ntextStatus : " + textStatus + "\nerrorThrown : " + errorThrown);
            }
          });
    });
});

});

});

