require (['require-config'], function () {

require(['jquery','js/left-nav', 'js/top-nav'], function($, leftnav, topnav) {

$(document).ready(function () {
  topnav.loadTopNav();
  leftnav.loadLeftNav("left-nav-dash");
});

});

});

