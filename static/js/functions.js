require (['require-config'], function () {

require(['jquery','js/left-nav', 'js/top-nav', 'bootstrap', 'bootstrap-table', 'bootstrap-table-local'], function($, leftnav, topnav) {

function loadFunctionList () {
  var funcTable = $('#tbl-functions')
  $.ajax({
      type: 'GET',
      //url: "/data/functions_mock.json",
      url: "/functions/?Marker=0&MaxItems=100",
      dataType: "json",
      success: function(resultData) {
        $.extend($.fn.bootstrapTable.defaults, $.fn.bootstrapTable.locales['en-US']);
        funcTable.bootstrapTable({
            clickToSelect: true,
            columns: [
            {
                field: 'state',
                radio	: true,
                align: 'center',
                valign: 'middle',
                clickToSelect: false,
            }, {
                field: 'FunctionName',
                title: 'Function Name',
                sortable: true,
            }, {
                field: 'Description',
                title: 'Description',
                sortable: true,
            }, {
                field: 'Runtime',
                title: 'Runtime',
                sortable: true,
            }, {
                field: 'CodeSize',
                title: 'Code Size',
                sortable: true,
            }, {
                field: 'LastModified',
                title: 'LastModified',
                sortable: true,
            }],
            data: resultData.Functions
        });

        funcTable.on ('click-row.bs.table', function (e, row, $element) {
            url = "/function.html?name=" + row.FunctionName;
            window.location.href = url;
        });
      },
      failure: function(errMsg) {
            alert(errMsg);
      },
      error: function(e) {
        alert(e);
      }
    });

    function getIdSelections() {
        return $.map(funcTable.bootstrapTable('getSelections'), function (row) {
            return row.FunctionName;
        });
    }

    $("#btn-func-test").click(function () {
        var funcs = getIdSelections();
        if (funcs.length == 0){
             alert("you must select a function!");
        }
        else{
            url = "/function.html?name=" + funcs[0] + "&eventModal=open";
            window.location.href = url;
        }
    });

    $("#btn-func-del").click(function () {
        var funcs = getIdSelections();

        if (funcs.length == 0){
             alert("you must select a function!");
        }
        else{
            $.ajax({
                type: 'DELETE',
                url: "/functions/" + funcs[0],
                dataType: "json",
                success: function(resultData) {
                  // delete function
                  alert("Delete function : " + funcs[0]);

                  // reload table
                  $.ajax({
                      type: 'GET',
                      url: "/functions/?Marker=0&MaxItems=20",
                      dataType: "json",
                      success: function(resultData) {
                        funcTable.bootstrapTable("load", resultData.Functions);
                      },
                      failure: function(errMsg) {
                            alert(errMsg);
                      },
                      error: function(e) {
                        alert(e);
                      }
                    });
                },
                failure: function(errMsg) {
                      alert(errMsg);
                },
                error: function(e) {
                  alert(e);
                }
              });
      }
    });
}

$(document).ready(function () {
  topnav.loadTopNav();
  leftnav.loadLeftNav("left-nav-func");
  loadFunctionList ();

  $("#btn-func-create").click (function () {
    window.location.href = "createfunc.html";
  });
});

});

});

