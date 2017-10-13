angular.module('mmumshad.yamljs', []).provider('YAML', function () {
  // paste minified d3 code in here
  this.$get = ['$window', function ($window) {
    // configure JSONEditor using provider's configuration
    return $window.YAML;
  }];

});
