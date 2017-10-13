# angular-yamljs

This is an angular wrapper for [YAMLJS](https://github.com/jeremyfa/yaml.js) written by [Jeremy Faivre](https://github.com/jeremyfa)


How to Install
----------

``` bash
bower install angular-yamljs
```

How to use
----------
Add dependency to angular application

``` js
angular.module('app', ['mmumshad.yamljs']);
```

Use it in code

``` js
angular.module('app')
  .controller('Ctrl1', function (YAML) {
    // parse YAML string
    nativeObject = YAML.parse(yamlString);
    
    // Generate YAML
    yamlString = YAML.stringify(nativeObject, 4);
    
    // Load yaml file using YAML.load
    nativeObject = YAML.load('myfile.yml');
  });
```
