require (['require-config'], function () {

require(['jquery','js/left-nav', 'js/top-nav', 'bootstrap', 'bootstrap-table', 'bootstrap-table-local'], function($, leftnav, topnav) {

function loadTopicList () {
  var topicTable = $('#tbl-topics')
  $.ajax({
      type: 'GET',
      url: "/data/functions_mock.json",
      dataType: "json",
      success: function(resultData) {
        $.extend($.fn.bootstrapTable.defaults, $.fn.bootstrapTable.locales['en-US']);
        topicTable.bootstrapTable({
            clickToSelect: true,
            columns: [
            {
                field: 'state',
                radio	: true,
                align: 'center',
                valign: 'middle',
                clickToSelect: false
            }, {
                field: 'FunctionName',
                title: 'Topic name',
                sortable: true,
            }, {
                field: 'Description',
                title: 'Description',
                sortable: true,
            }],
            data: resultData.Functions
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
        return $.map(topicTable.bootstrapTable('getSelections'), function (row) {
            return row.FunctionName;
        });
    }
}

$(document).ready(function () {
  topnav.loadTopNav();
  leftnav.loadLeftNav("left-nav-kafka");
  loadTopicList ();
});

});

});

