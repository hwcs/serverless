
define(function() {

var loadLeftNav = function (actived) {
  $("#left-nav").load("left-nav.html",function(responseTxt,statusTxt,xhr){
    if(statusTxt=="success") {
      $("#"+actived).attr("class", "active")
    }
  });
}

return {
  loadLeftNav : loadLeftNav
};

});

