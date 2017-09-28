# angular-swagger-ui

`angular-swagger-ui` is an angularJS implementation of [Swagger UI](http://swagger.io)

Swagger helps you documenting your RESTful API.

Swagger UI helps developers discovering your RESTful API by providing an online documentation with an integrated API explorer.

### Warning 
> By default, only JSON Swagger 2.0 is supported.
To handle Swagger 1.2 please add module `swagger1-to-swagger2-converter` see [Enable Swagger 1.2](#enable-swagger-12).
To handle YAML please add module `swagger-yaml-parser` see [Enable YAML](#enable-yaml)

> By default, Authorization is not supported.
To handle authorization (oauth2 is not implemented) please add module `swagger-auth` see [Enable authorization](#enable-authorization)

## Demo

A sample app using `angular-swagger-ui` is available here:

http://orange-opensource.github.io/angular-swagger-ui

## Quick Start

### Install

`bower install angular-swagger-ui --save`

### Dependencies

1. [angularJS](https://angularjs.org)
2. [bootstrap CSS](http://getbootstrap.com)
3. [angular-ui-bootstrap](https://angular-ui.github.io/bootstrap/) (required only if using [Authorization](#enable-authorization))

## License

All code in this repository is covered by the [MIT license](http://opensource.org/licenses/MIT).
See LICENSE file for copyright details.

## Getting Started

Include `angular-swagger-ui` as a dependency into your application

As some properties of Swagger specifications can be formatted as HTML:

* You **SHOULD** include `ngSanitize` as a dependency into your application (avoids JS injection) if Swagger specifications are loaded from **untrusted** sources (see `dist/index.html` as an example)
* You **CAN** add `trusted-sources="true"` as directive parameter (avoids embedding `ngSanitize`) if Swagger specifications are loaded from **trusted** sources (see `src/index.html` as an example)
* You **MUST** at least choose one of the two previous solutions

```html
<script type="text/javascript">
	// If directive has parameter trusted-sources="true"
	angular.module('yourApp', ['swaggerUi']);
	...
	// OR if you choosed to use "ngSanitize"
	angular.module('yourApp', ['ngSanitize', 'swaggerUi']);
	...
</script>
```
Create an HTML element in your angularJS application's template or in your HTML page
```html
<div swagger-ui url="URLToYourSwaggerSpecification" api-explorer="true"></div>
```
Add `swagger-ui.min.js` and `angular.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<!-- if you choosed to use "ngSanitize" -->
 	<script src="yourPathToAngularSanitize/angular-sanitize.min.js"></script>
</body>
```
Add `swagger-ui.min.css` and `bootstrap.min.css` to the head of the HTML page.
```html
<body>
	<head>
		...
		<link rel="stylesheet" href="yourPathToBootstrapCSS/bootstrap.min.css">
		<link rel="stylesheet" href="yourPathToAngularSwaggerUI/dist/css/swagger-ui.min.css">
  	</head>
</body>
```

## Parameters

#### API explorer
Display or not API explorer, default is `false`
```html
<div swagger-ui url="URLToYourSwaggerSpecification" api-explorer="true/false"></div>
```

#### Swagger specification loading indicator
`yourScopeVariable` will be assigned to `true` or `false` depending on Swagger specification loading status
```html
<div ng-show="yourScopeVariable">loading ...</div>
<div swagger-ui url="URLToYourSwaggerSpecification" loading="yourScopeVariable"></div>
```

#### Error handler
Define an error handler to catch errors, if none defined `console.error` is used
```html
<div swagger-ui url="URLToYourSwaggerSpecification" error-handler="yourErrorHandler"></div>
```
```js
$scope.yourErrorHandler = function(/*String or Object*/ message, /*Integer*/ code){
	
}
```

#### Permalinks
Allows having a URL direct access to a group of operations or to an operation and making it unfolded at startup
```html
<div swagger-ui url="URLToYourSwaggerSpecification" permalinks="true/false"></div>
```

#### Swagger validator
Disable Swagger validator or define a custom Swagger validator.
If parameter not defined, the validator will be 'http://online.swagger.io/validator'
```html
<div swagger-ui url="URLToYourSwaggerSpecification" validator-url="false or URL"></div>
```

#### Parser type
Swagger specification parser is chosen depending on the `Content-Type` of the specification response. If host serving your Swagger specification does not send `Content-Type: application/json` then you can force the parser to JSON:
```html
<div swagger-ui url="URLToYourSwaggerSpecification" parser="json"></div>
```

#### Template URL
Define a custom template to be used by SwaggerUI
```html
<div swagger-ui url="URLToYourSwaggerSpecification" template-url="yourTemplatePath"></div>
```

#### Input type and input
##### Render a Swagger specification from JSON object
```html
<div swagger-ui input-type="json" input="yourJsonObject"></div>
```

##### Render a Swagger specification from YAML string
Make sure to use module `swagger-yaml-parser`, see [Enable YAML](#enable-yaml)
```html
<div swagger-ui input-type="yaml" input="yourYamlString"></div>
```

##### Render a Swagger specification from URL (same behavior as using "url" parameter)
```html
<div swagger-ui input-type="url" input="yourURL"></div>
```

## i18n

#### Built-in languages
`angular-swagger-ui` is available in english and french, english is used by default

To use french, add `fr.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/i18/fr.min.js"></script>
</body>
```
Set language to french at startup
```html
<script type="text/javascript">
	angular
		.module('yourApp', ['swaggerUi'])
		.config(function(swaggerTranslatorProvider) {
			swaggerTranslatorProvider.setLanguage('fr');
		});
	...
</script>
```
Set language to french at runtime
```html
<script type="text/javascript">
	angular
		.module('yourApp', ['swaggerUi'])
		.controller('yourController', function(swaggerTranslator) {
			swaggerTranslator.useLanguage('fr');
		});
	...
</script>
```

#### Add languages
You can add your own languages, see `src/scripts/i18n/en.js` to find the keys you have to override
```html
<script type="text/javascript">
	angular
		.module('yourApp', ['swaggerUi'])
		.config(function(swaggerTranslatorProvider) {
			swaggerTranslatorProvider.addTranslations('yourLangId', {
				key: 'value'
				...
			});
			swaggerTranslatorProvider.setLanguage('yourLangId');
		});
	...
</script>
```

#### Internationalize your app
You can also use `swaggerTranslator` to internationalize your app by using a service, a directive or a filter
```html
<body>
 	...
 	<div swagger-translate="yourKey" swagger-translate-value="yourParam"></div>
 	<div ng-bind="yourDynamicKey|swaggerTranslate:yourDynamicParam"></div>
 	...
	<script type="text/javascript">
		angular
			.module('yourApp', ['swaggerUi'])
			.config(function(swaggerTranslatorProvider) {
				swaggerTranslatorProvider.addTranslations('en', {
					yourKey: 'blablabla {{propertyNameOfYourParam}}'
					...
				});
			})
			.controller('yourController', function(swaggerTranslator) {
				var localizedMessage = swaggerTranslator.translate('yourKey', yourParam);
			});
		...
	</script>
</body>
```

## Customization

#### Enable authorization
`oauth` is not implemented, only `basic` and `API key` authorizations are implemented.
Add `swagger-auth.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-auth.min.js"></script><!-- without angular-ui-bootstrap modal embedded -->
 	OR
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-auth-ui-boostrap-modal.min.js"></script><!-- angular-ui-bootstrap modal embedded -->
 	...
	<script type="text/javascript">
		angular
			.module('yourApp', ['swaggerUi', 'swaggerUiAuthorization'])
			...
	</script>
</body>
```

#### Enable Swagger 1.2
Add `swagger1-to-swagger2-converter.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger1-to-swagger2-converter.min.js"></script>
</body>
```

#### Enable Swagger external references
See [Swagger 2.0 spec](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#relative-schema-file-example).
Add `swagger-external-references.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-external-references.min.js"></script>
</body>
```

#### Enable XML formatter on API explorer responses
Add `swagger-xml-formatter.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-xml-formatter.min.js"></script>
</body>
```

#### Enable YAML
Add [js-yaml library](https://cdnjs.com/libraries/js-yaml).
Add `swagger-yaml-parser.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToJsYaml/js-yaml.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-yaml-parser.min.js"></script>
</body>
```

#### Enable markdown
Add [marked library](https://cdnjs.com/libraries/marked).
Add `swagger-markdown.min.js` at the end of the body
```html
<body>
 	...
 	<script src="yourPathToAngularJS/angular.min.js"></script>
 	<script src="yourPathToMarked/marked.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/swagger-ui.min.js"></script>
 	<script src="yourPathToAngularSwaggerUI/dist/scripts/modules/swagger-markdown.min.js"></script>
</body>
```

#### Writing your own modules
Modifying `angular-swagger-ui` can be achieved by writing your own modules. As an example your can have a look at the ones in `src/scripts/modules`.
A module is an object (can be a service) having a function `execute` which must return a promise.

You can make your module modifying behaviours at different phases:

* `BEFORE_LOAD`: allows modifying Swagger specification request before it is sent
* `BEFORE_PARSE`: allows modifying Swagger specification after it has been loaded
* `PARSE`: allows adding a Swagger parser for content types other than JSON
* `BEFORE_DISPLAY`: allows modifying internal parsed Swagger specification before it is displayed
* `BEFORE_EXPLORER_LOAD`: allows modifying API explorer request before it is sent
* `AFTER_EXPLORER_LOAD`: allows modifying API explorer response before it is displayed

```js
angular
	.module('myApp', ['swaggerUi'])
	.service('myModule', function($q) {

		this.execute = function() {
			var deferred = $q.defer();
			// if nothing done: call deferred.resolve(false);
			// if success: call deferred.resolve(true);
			// if error: call deferred.reject({message: 'error message', code: 'error_code'});
			return deferred.promise;
		}

	})
	.run(function(swaggerModules, myModule){
		swaggerModules.add(swaggerModules.BEFORE_LOAD, myModule);
	})
	...

```
