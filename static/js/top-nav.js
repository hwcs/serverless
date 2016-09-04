define(function() {

var loadTopNav = function () {
  $("#top-nav").load("top-nav.html",function(responseTxt,statusTxt,xhr){
    if(statusTxt=="success") {
      $("#top-nav").attr("class", "container-fluid");
    }
  });
}

return {
  loadTopNav : loadTopNav
};

});

