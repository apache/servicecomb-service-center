/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi', ['ng'])
	.directive('swaggerUi', ["$injector", function($injector) {

		return {
			restrict: 'A',
			controller: 'swaggerUiController',
			templateUrl: function(element, attrs) {
				return attrs.templateUrl || 'templates/swagger-ui.html';
			},
			scope: {
				// Swagger specification URL (string, required)
				url: '=?',
				// Swagger specification parser type (string, optional, default = "auto")
				// Built-in allowed values:
				// 		"auto": (default) parser is based on response Content-Type
				//		"json": force using JSON parser
				//		"yaml": force using YAML parser, requires to use module 'swagger-yaml-parser'
				//
				// More types could be defined by external modules
				parser: '@?',
				// Swagger specification loading indicator (variables, optional)
				loading: '=?',
				// Use permalinks? (boolean, optional, default = false)
				permalinks: '=?',
				// Display API explorer (boolean, optional, default = false)
				apiExplorer: '=?',
				// Error handler (function, optional)
				errorHandler: '=?',
				// Are Swagger specifications loaded from trusted source only ? (boolean, optional, default = false)
				// If true, it avoids using ngSanitize but consider HTML as trusted so won't be cleaned
				trustedSources: '=?',
				// Allows defining a custom Swagger validator or disabling Swagger validation
				// If false, Swagger validation will be disabled
				// If URL, will be used as Swagger validator
				// If not defined, validator will be 'http://online.swagger.io/validator'
				validatorUrl: '@?',
				// Specifies the type of "input" parameter to allow rendering Swagger specification from object or string (string, optional)
				// Allowed values:
				// 		"url": (default) "input" parameter is an URL
				//		"json": "input" parameter is a JSON object
				//		"yaml": "input" parameter is a YAML string, requires to use module 'swagger-yaml-parser'
				//
				inputType: '@?',
				// Allows rendering an external Swagger specification (string or object, optional)
				input: '=?'
			},
			link: function(scope) {
				// check parameters
				if (!scope.trustedSources && !$injector.has('$sanitize')) {
					console.warn('AngularSwaggerUI: You must use ngSanitize OR set trusted-sources=true as directive param if swagger specifications are loaded from trusted sources');
				}
				if (scope.validatorUrl === undefined) {
					scope.validatorUrl = 'http://online.swagger.io/validator';
				}
			}
		};
	}])
	.controller('swaggerUiController', ["$scope", "$window", "$http", "$location", "$anchorScroll", "$timeout", "swaggerClient", "swaggerModules", "swaggerTranslator", function($scope, $window, $http, $location, $anchorScroll, $timeout, swaggerClient, swaggerModules, swaggerTranslator) {

		var swagger;

		// WARNING authentication is not implemented, please use modules to customize API calls, see README.md

		/**
		 * Load Swagger specification
		 */
		function loadSwagger(url, callback) {
			$scope.loading = true;
			var options = {
				method: 'GET',
				url: url
			};
			swaggerModules
				.execute(swaggerModules.BEFORE_LOAD, options)
				.then(function() {
					$http(options)
						.then(callback)
						.catch(function(response) {
							onError({
								code: response.status,
								message: response.data
							});
						});
				})
				.catch(onError);
		}

		/**
		 * Swagger specification has been loaded, launch parsing
		 */
		function swaggerLoaded(swaggerUrl, swaggerType) {
			$scope.loading = false;
			var parseResult = {};
			// execute modules
			$scope.parser = $scope.parser || 'auto';
			swaggerModules
				.execute(swaggerModules.PARSE, $scope.parser, swaggerUrl, swaggerType, swagger, $scope.trustedSources, parseResult)
				.then(function(executed) {
					if (executed) {
						if (parseResult.transformSwagger) {
							swagger = parseResult.transformSwagger;
							delete parseResult.transformSwagger;
						}
						swaggerParsed(parseResult);
					} else {
						onError({
							code: 415,
							message: swaggerTranslator.translate('errorNoParserFound', {
								type: swaggerType,
								version: swagger.swagger
							})
						});
					}
				})
				.catch(onError);
		}

		/**
		 * Swagger specification has parsed, launch display
		 */
		function swaggerParsed(parseResult) {
			// execute modules
			swaggerModules
				.execute(swaggerModules.BEFORE_DISPLAY, parseResult)
				.then(function() {
					// display swagger UI
					$scope.infos = parseResult.infos;
					$scope.form = parseResult.form;
					$scope.resources = parseResult.resources;
					if ($scope.permalinks) {
						$timeout(function() {
							$anchorScroll();
						}, 200);
					}
				})
				.catch(onError);
		}

		function onError(error) {
			$scope.loading = false;
			if (typeof $scope.errorHandler === 'function') {
				$scope.errorHandler(error.message, error.code);
			} else {
				console.error(error.code, 'AngularSwaggerUI: ' + error.message);
			}
		}

		function watchData() {
			$scope.$watch('input', function(data) {
				//reset
				$scope.infos = {};
				$scope.resources = [];
				$scope.form = {};
				if (data) {
					swagger = angular.copy(data);
					swaggerLoaded(null, 'application/' + $scope.inputType);
				}
			});
		}

		function watchUrl(key) {
			$scope.$watch(key, function(url) {
				//reset
				$scope.infos = {};
				$scope.resources = [];
				$scope.form = {};
				if (url && url !== '') {
					if ($scope.loading) {
						//TODO cancel current loading swagger
					}
					if ($scope.validatorUrl && url.indexOf('http') !== 0) {
						// make URL absolute to make validator working
						$scope.url = absoluteUrl(url);
						return;
					}
					// load Swagger specification
					loadSwagger(url, function(response) {
						swagger = response.data;
						// execute modules
						swaggerModules
							.execute(swaggerModules.BEFORE_PARSE, url, swagger)
							.then(function() {
								var contentType = response.headers()['content-type'] || 'application/json',
									swaggerType = contentType.split(';')[0];

								swaggerLoaded(url, swaggerType);
							})
							.catch(onError);
					});
				}
			});
		}

		switch ($scope.inputType) {
			case 'json':
			case 'yaml':
				$scope.validatorUrl = false; // disable validator
				watchData();
				break;
			case 'url':
				watchUrl('input');
				break;
			default:
				watchUrl('url');
				break;
		}

		/**
		 * transform a relative URL to an absolute URL
		 */
		function absoluteUrl(url) {
			var a = document.createElement('a');
			a.href = url;
			return a.href;
		}

		/**
		 * show all resource's operations as list or as expanded list
		 */
		$scope.expand = function(resource, expandOperations) {
			resource.open = true;
			for (var i = 0, op = resource.operations, l = op.length; i < l; i++) {
				op[i].open = expandOperations;
			}
		};

		$window.swaggerlink = $scope.permalink = function(name) {
			if ($scope.permalinks) {
				$location.hash(name);
			}
			$timeout(function() {
				$anchorScroll(name);
			}, 50);
		};

		/**
		 * sends a sample API request
		 */
		$scope.submitExplorer = function(operation) {
			operation.loading = true;
			swaggerClient
				.send(swagger, operation, $scope.form[operation.id])
				.then(function(result) {
					operation.loading = false;
					operation.explorerResult = result;
				});
		};

		/**
		 * handle operation's authentication params
		 */
		$scope.auth = function(operation) {
			var i = 0,
				sec, key, auth = [],
				security = operation.security;

			for (; i < security.length; i++) {
				sec = security[i];
				for (key in sec) {
					auth.push(swagger.securityDefinitions[key]);
				}
			}
			swaggerModules
				.execute(swaggerModules.AUTH, operation, auth)
				.catch(onError);
		};

		/**
		 * check if operation's authorization params are set
		 */
		$scope.authValid = function(operation) {
			var i = 0,
				sec, auth, key,
				security = operation.security;

			for (; i < security.length; i++) {
				sec = security[i];
				for (key in sec) {
					auth = swagger.securityDefinitions[key];
					if (auth.valid) {
						operation.authParams = angular.copy(auth);
						operation.authParams.scopes = sec[key];
						return true;
					}
				}
			}
			return false;
		};

	}])
	.directive('fileInput', function() {
		// helper to be able to retrieve HTML5 File in ngModel from input
		return {
			restrict: 'A',
			require: 'ngModel',
			link: function(scope, element, attr, ngModel) {
				element.bind('change', function() {
					scope.$apply(function() {
						//TODO manage multiple files ?
						ngModel.$setViewValue(element[0].files[0]);
					});
				});
			}
		};
	});
/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.service('swaggerClient', ["$q", "$http", "$httpParamSerializer", "swaggerModules", function($q, $http, $httpParamSerializer, swaggerModules) {

		/**
		 * format API explorer response before display
		 */
		function formatResult(deferred, response) {
			var query = '',
				data = response.data,
				config = response.config;

			if (config.params) {
				var parts = [];
				for (var key in config.params) {
					parts.push(key + '=' + encodeURIComponent(config.params[key]));
				}
				if (parts.length > 0) {
					query = '?' + parts.join('&');
				}
			}
			deferred.resolve({
				url: config.url + query,
				response: {
					body: data ? (angular.isString(data) ? data : angular.toJson(data, true)) : 'no content',
					status: response.status,
					headers: angular.toJson(response.headers(), true)
				}
			});
		}

		/**
		 * Send API explorer request
		 */
		this.send = function(swagger, operation, values) {
			var deferred = $q.defer(),
				query = {},
				headers = {},
				path = operation.path,
				urlEncoded = values.contentType === 'application/x-www-form-urlencoded',
				body;

			// build request parameters
			for (var i = 0, params = operation.parameters || [], l = params.length; i < l; i++) {
				//TODO manage 'collectionFormat' (csv etc.) !!
				var param = params[i],
					value = values[param.name];

				switch (param.in) {
					case 'query':
						if (!!value) {
							query[param.name] = value;
						}
						break;
					case 'path':
						path = path.replace(new RegExp('{' + param.name + '[^}]*}'), encodeURIComponent(value));
						break;
					case 'header':
						if (!!value) {
							headers[param.name] = value;
						}
						break;
					case 'formData':
						body = body || (urlEncoded ? {} : new FormData());
						if (!!value) {
							if (param.type === 'file') {
								values.contentType = undefined; // make browser defining it by himself
							}
							if (urlEncoded) {
								body[param.name] = value;
							} else {
								body.append(param.name, value);
							}
						}
						break;
					case 'body':
						body = body || value;
						break;
				}
			}

			// authorization
			var authParams = operation.authParams;
			if (authParams) {
				switch (authParams.type) {
					case 'apiKey':
						switch (authParams.in) {
							case 'header':
								headers[authParams.name] = authParams.apiKey;
								break;
							case 'query':
								query[authParams.name] = authParams.apiKey;
								break;
						}
						break;
					case 'basic':
						headers.Authorization = 'Basic ' + btoa(authParams.login + ':' + authParams.password);
						break;
				}
			}

			// add headers
			headers.Accept = values.responseType;
			headers['Content-Type'] = body ? values.contentType : 'text/plain';
			headers['X-InstanceIP'] = swagger.instanceIP || '';
			headers['X-SchemaName'] = swagger.schemaName || '';
		    // build request
			var basePath = swagger.basePath || '',
				baseUrl = [
					swagger.schemes[0],
					'://',
					swagger.host,
					basePath.length > 0 && basePath.substring(basePath.length - 1) === '/' ? basePath.slice(0, -1) : basePath
				].join(''),
				options = {
					method: operation.httpMethod,
					url: baseUrl + path,
					headers: headers,
					data: urlEncoded ? $httpParamSerializer(body) : body,
					params: query
				},
				callback = function(result) {
					// execute modules
					var response = {
						data: result.data,
						status: result.status,
						headers: result.headers,
						config: result.config
					};
					swaggerModules
						.execute(swaggerModules.AFTER_EXPLORER_LOAD, response)
						.then(function() {
							formatResult(deferred, response);
						});
				};

			// execute modules
			swaggerModules
				.execute(swaggerModules.BEFORE_EXPLORER_LOAD, options)
				.then(function() {
					// send request
					$http(options)
						.then(callback)
						.catch(callback);
				});

			return deferred.promise;
		};

	}]);

/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.service('swaggerModel', ["swaggerTranslator", function(swaggerTranslator) {

		var INLINE_MODEL_NAME = 'InlineModel';

		/**
		 * sample object cache to avoid generating the same one multiple times
		 */
		var sampleCache = {};

		/**
		 * retrieves object definition
		 */
		var resolveReference = this.resolveReference = function(swagger, object) {
			if (object.$ref) {
				var parts = object.$ref.replace('#/', '').split('/');
				object = swagger;
				for (var i = 0, j = parts.length; i < j; i++) {
					object = object[parts[i]];
				}
			}
			return object;
		};

		/**
		 * determines a property type
		 */
		var getType = this.getType = function(item) {
			var format = item.format;
			switch (format) {
				case 'int32':
					format = item.type;
					break;
				case 'int64':
					format = 'long';
					break;
			}
			return format || item.type;
		};

		/**
		 * handles allOf property of a schema
		 */
		function resolveAllOf(swagger, schema) {
			if (schema.allOf) {
				angular.forEach(schema.allOf, function(def) {
					angular.merge(schema, resolveReference(swagger, def));
				});
				delete schema.allOf;
			}
			return schema;
		}

		/**
		 * generates a sample object (request body or response body)
		 */
		function getSampleObj(swagger, schema, currentGenerated) {
			var sample, def, name, prop;
			currentGenerated = currentGenerated || {}; // used to handle circular references
			schema = resolveAllOf(swagger, schema);
			if (schema.default || schema.example) {
				sample = schema.default || schema.example;
			} else if (schema.properties) {
				sample = {};
				for (name in schema.properties) {
					prop = schema.properties[name];
					sample[name] = getSampleObj(swagger, prop.schema || prop, currentGenerated);
				}
			} else if (schema.additionalProperties) {
				// this is a map/dictionary
				// @see https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#model-with-mapdictionary-properties
				def = resolveReference(swagger, schema.additionalProperties);
				sample = {
					'string': getSampleObj(swagger, def, currentGenerated)
				};
			} else if (schema.$ref) {
				// complex object
				def = resolveReference(swagger, schema);
				if (def) {
					if (!sampleCache[schema.$ref] && !currentGenerated[schema.$ref]) {
						// object not in cache
						currentGenerated[schema.$ref] = true;
						sampleCache[schema.$ref] = getSampleObj(swagger, def, currentGenerated);
					}
					sample = sampleCache[schema.$ref] || {};
				} else {
					console.warn('schema not found', schema.$ref);
					sample = schema.$ref;
				}
			} else if (schema.type === 'array') {
				sample = [getSampleObj(swagger, schema.items, currentGenerated)];
			} else if (schema.type === 'object') {
				sample = {};
			} else {
				sample = schema.defaultValue || schema.example || getSampleValue(schema);
			}
			return sample;
		}

		/**
		 * generates a sample value for a basic type
		 */
		function getSampleValue(schema) {
			var result,
				type = getType(schema);

			switch (type) {
				case 'long':
				case 'integer':
					result = 0;
					break;
				case 'boolean':
					result = false;
					break;
				case 'float':
				case 'double':
				case 'number':
					result = 0.0;
					break;
				case 'string':
				case 'byte':
				case 'binary':
				case 'password':
					result = 'string';
					if (schema.enum && schema.enum.length > 0) {
						result = schema.enum[0];
					}
					break;
				case 'date':
					result = (new Date()).toISOString().split('T')[0];
					break;
				case 'date-time':
					result = (new Date()).toISOString();
					break;
			}
			return result;
		}

		/**
		 * generates a sample JSON string (request body or response body)
		 */
		this.generateSampleJson = function(swagger, schema) {
			try {
				var json,
					obj = getSampleObj(swagger, schema);

				if (obj) {
					json = angular.toJson(obj, true);
				}
				return json;

			} catch (ex) {
				console.error('failed to generate sample json', schema, ex);
				return 'failed to generate sample json';
			}
		};

		/**
		 * retrieves object class name based on $ref
		 */
		function getClassName(schema) {
			var parts = schema.$ref.split('/');
			return parts[parts.length - 1];
		}

		/**
		 * inline model counter
		 */
		var countInLineModels = 1;

		/**
		 * is model attribute required ?
		 */
		function isRequired(item, name) {
			return item.required && item.required.indexOf(name) !== -1;
		}

		/**
		 * generates new inline model name
		 */
		function getInlineModelName() {
			var name = INLINE_MODEL_NAME + (countInLineModels++);
			return name;
		}

		/**
		 * model object caches to avoid generating the same one multiple times
		 */
		var modelCache = {};
		var modelCacheIds = {};

		/**
		 * generate a model and its submodels from schema
		 */
		this.generateModel = function(swagger, schema, operationId) {
			try {
				schema = resolveAllOf(swagger, schema);
				var model = [],
					subModelIds = {},
					subModels = {};

				if (schema.properties) {
					// if inline model
					subModels[getInlineModelName()] = schema;
					subModels = angular.merge(subModels, findAllModels(swagger, schema, subModelIds));
				} else {
					subModels = findAllModels(swagger, schema, subModelIds);
				}

				if (!schema.$ref && !schema.properties) {
					// if array or map/dictionary or simple type
					model.push('<strong>', getModelProperty(schema, subModels, subModelIds, operationId), '</strong><br><br>');
				}
				angular.forEach(subModels, function(schema, modelName) {
					model.push(getModelMaybeFromCache(swagger, schema, modelName, subModels, subModelIds, operationId));
				});
				return model.join('');

			} catch (ex) {
				console.error('failed to generate model', schema, ex);
				return 'failed to generate model';
			}
		};

		/**
		 * find all models to generate
		 */
		function findAllModels(swagger, schema, subModelIds, modelName, onGoing) {
			var models = {};
			if (modelName) {
				onGoing = onGoing || {}; // used to handle circular definitions
				if (onGoing[modelName]) {
					return models;
				}
				onGoing[modelName] = true;
			}
			schema = resolveAllOf(swagger, schema);
			if (schema.properties) {
				angular.forEach(schema.properties, function(property) {
					inspectSubModel(swagger, property, models, subModelIds, onGoing);
				});
			} else if (schema.schema || schema.$ref) {
				var subSchema = schema.schema || schema,
					def = resolveReference(swagger, subSchema),
					subPropertyModelName = getClassName(subSchema);

				if (def) {
					models[subPropertyModelName] = def;
					subModelIds[subPropertyModelName] = countModel++;
					angular.merge(models, findAllModels(swagger, def, subModelIds, subPropertyModelName, onGoing));
				}
			} else if (schema.type === 'array') {
				inspectSubModel(swagger, schema.items, models, subModelIds, onGoing);
			} else if (schema.additionalProperties) {
				// this is a map/dictionary
				inspectSubModel(swagger, schema.additionalProperties, models, subModelIds, onGoing);
			}
			return models;
		}

		/**
		 * look for submodels
		 */
		function inspectSubModel(swagger, schema, models, subModelIds, onGoing) {
			var inlineModelName = generateInlineModel(schema, models, subModelIds);
			angular.merge(models, findAllModels(swagger, schema, subModelIds, inlineModelName, onGoing));
		}

		/**
		 * generates an inline model if needed
		 */
		function generateInlineModel(subProperty, models, subModelIds) {
			var subModelName;
			if (subProperty.properties) {
				subModelName = getInlineModelName();
				subProperty.modelName = subModelName;
				subModelIds[subModelName] = countModel++;
				models[subModelName] = subProperty;
			}
			return subModelName;
		}

		/**
		 * get model from cache or generate it
		 */
		function getModelMaybeFromCache(swagger, schema, modelName, subModels, subModelIds, operationId) {
			var model,
				useCache = false;

			if (modelName.indexOf(INLINE_MODEL_NAME) === -1) {
				// inline models are not cached
				useCache = true;
			}
			if (useCache) {
				if (modelCache[modelName]) {
					model = modelCache[modelName];
					// substitute moel IDs
					angular.forEach(modelCacheIds, function(id, subName) {
						model = model.replace(id, operationId + '-model-' + subModelIds[subName]);
					});
				} else {
					model = modelCache[modelName] = getModel(swagger, schema, modelName, subModels, subModelIds, operationId);
					modelCacheIds[modelName] = operationId + '-model-' + subModelIds[modelName];
				}
			} else {
				model = getModel(swagger, schema, modelName, subModels, subModelIds, operationId);
			}
			return model;
		}

		/**
		 * generates an HTML link to a submodel
		 */
		function getSubModelLink(operationId, modelId, name) {
			return ['<a class="model-link type" onclick="swaggerlink(\'', operationId, '-model-', modelId, '\')">', name, '</a>'].join('');
		}

		/**
		 * model counter
		 */
		var countModel = 0;

		/**
		 * generates a single model in HTML
		 */
		function getModel(swagger, schema, modelName, subModels, subModelIds, operationId) {
			var buffer = ['<div class="model" id="', operationId + '-model-' + subModelIds[modelName], '">'];
			if (schema.properties) {
				buffer.push('<div><strong>' + modelName + ' {</strong></div>');
				var hasProperties = false;
				angular.forEach(schema.properties, function(property, propertyName) {
					hasProperties = true;
					buffer.push('<div class="pad"><strong>', propertyName, '</strong> (<span class="type">');
					buffer.push(getModelProperty(property, subModels, subModelIds, operationId));
					buffer.push('</span>');
					if (!isRequired(schema, propertyName)) {
						buffer.push(', ', '<em>' + swaggerTranslator.translate('modelOptional') + '</em>');
					}
					buffer.push(')');
					if (property.description) {
						buffer.push(': ', property.description);
					}
					if (property.enum) {
						buffer.push(' = ', angular.toJson(property.enum).replace(/,/g, swaggerTranslator.translate('modelOr')));
					}
					buffer.push(',</div>');
				});
				if (hasProperties) {
					buffer.pop();
					buffer.push('</div>');
				}
				buffer.push('<div><strong>}</strong></div>');
			} else if (schema.type === 'array' || schema.additionalProperties) {
				buffer.push(getModelProperty(schema, subModels, subModelIds, operationId));
			} else if (schema.type === 'object') {
				buffer.push('<strong>', modelName || getInlineModelName(), ' {<br>}</strong>');
			} else if (schema.type) {
				buffer.push('<strong>', getType(schema), '</strong>');
			}
			return buffer.join('');
		}

		/**
		 * retrieves model property class
		 */
		function getModelProperty(property, subModels, subModelIds, operationId) {
			var modelName, buffer = [];
			if (property.modelName) {
				buffer.push(getSubModelLink(operationId, subModelIds[property.modelName], property.modelName));
			} else if (property.schema || property.$ref) {
				modelName = getClassName(property.schema || property);
				buffer.push(getSubModelLink(operationId, subModelIds[modelName], modelName));
			} else if (property.type === 'array') {
				buffer.push('Array[');
				buffer.push(getModelProperty(property.items, subModels, subModelIds, operationId));
				buffer.push(']');
			} else if (property.properties) {
				buffer.push(getSubModelLink(operationId, subModelIds[property.modelName], modelName));
			} else if (property.additionalProperties) {
				// this is a map/dictionary
				buffer.push('Map&lt;string, ');
				buffer.push(getModelProperty(property.additionalProperties, subModels, subModelIds, operationId));
				buffer.push('&gt;');
			} else {
				buffer.push(getType(property));
			}
			return buffer.join('');
		}

		/**
		 * clears generated samples cache
		 */
		this.clearCache = function() {
			sampleCache = {};
			modelCache = {};
			modelCacheIds = {};
			countModel = 0;
			countInLineModels = 1;
		};

	}]);
/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.service('swaggerModules', ["$q", function($q) {

		var modules = {};

		this.AUTH = 'AUTH';
		this.BEFORE_LOAD = 'BEFORE_LOAD';
		this.BEFORE_PARSE = 'BEFORE_PARSE';
		this.PARSE = 'PARSE';
		this.BEFORE_DISPLAY = 'BEFORE_DISPLAY';
		this.BEFORE_EXPLORER_LOAD = 'BEFORE_EXPLORER_LOAD';
		this.AFTER_EXPLORER_LOAD = 'AFTER_EXPLORER_LOAD';
		this.BEFORE_CONVERT = 'BEFORE_CONVERT';

		/**
		 * Adds a new module to swagger-ui
		 */
		this.add = function(phase, module) {
			if (!modules[phase]) {
				modules[phase] = [];
			}
			if (modules[phase].indexOf(module) < 0) {
				modules[phase].push(module);
			}
		};

		/**
		 * Runs modules' "execute" function one by one
		 */
		function executeAll(deferred, phaseModules, args, phaseExecuted) {
			var module = phaseModules.shift();
			if (module) {
				module
					.execute.apply(module, args)
					.then(function(executed) {
						phaseExecuted = phaseExecuted || executed;
						executeAll(deferred, phaseModules, args, phaseExecuted);
					})
					.catch(deferred.reject);
			} else {
				deferred.resolve(phaseExecuted);
			}
		}

		/**
		 * Executes modules' phase
		 */
		this.execute = function() {
			var args = Array.prototype.slice.call(arguments), // get an Array from arguments
				phase = args.splice(0, 1),
				deferred = $q.defer(),
				phaseModules = modules[phase] || [];

			executeAll(deferred, [].concat(phaseModules), args);
			return deferred.promise;
		};

	}]);

/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.service('swaggerParser', ["$q", "$sce", "$location", "swaggerModel", "swaggerTranslator", function($q, $sce, $location, swaggerModel, swaggerTranslator) {

		var trustedSources,
			operationId,
			paramId;

		/**
		 * Module entry point
		 */
		this.execute = function(parserType, url, contentType, data, isTrustedSources, parseResult) {
			var deferred = $q.defer();
			if (data.swagger === '2.0' && (parserType === 'json' || (parserType === 'auto' && contentType === 'application/json'))) {
				trustedSources = isTrustedSources;
				try {
					parseSwagger2Json(data, url, deferred, parseResult);
				} catch (e) {
					deferred.reject({
						code: 500,
						message: swaggerTranslator.translate('errorParseFailed', e)
					});
				}
			} else {
				deferred.resolve(false);
			}
			return deferred.promise;
		};

		/**
		 * parse swagger description to ease HTML generation
		 */
		function parseSwagger2Json(swagger, url, deferred, parseResult) {
			var map = {},
				form = {},
				resources = [],
				infos = swagger.info,
				openPath = $location.hash(),
				defaultContentType = 'application/json',
				sortResources = !swagger.tags;

			operationId = 0;
			paramId = 0;
			parseInfos(swagger, url, infos, defaultContentType);
			parseTags(swagger, resources, map);
			parseOperations(swagger, resources, form, map, defaultContentType, openPath);
			cleanUp(resources, openPath, sortResources);
			// prepare result
			parseResult.infos = infos;
			parseResult.resources = resources;
			parseResult.form = form;
			deferred.resolve(true);
		}

		/**
		 * parse main infos
		 */
		function parseInfos(swagger, url, infos, defaultContentType) {
			// build URL params
			var a = angular.element('<a href="' + url + '"></a>')[0];
			swagger.schemes = [swagger.schemes && swagger.schemes[0] || a.protocol.replace(':', '')];
			swagger.host = swagger.host || a.host;
			swagger.consumes = swagger.consumes || [defaultContentType];
			swagger.produces = swagger.produces || [defaultContentType];
			// build main infos
			infos.scheme = swagger.schemes[0];
			infos.basePath = swagger.basePath;
			infos.host = swagger.host;
			infos.description = trustHtml(infos.description);
			infos.externalDocs = swagger.externalDocs;
		}

		/**
		 * parse tags
		 */
		function parseTags(swagger, resources, map) {
			var i, l, tag;
			if (!swagger.tags) {
				resources.push({
					name: 'default',
					open: true
				});
				map['default'] = 0;
			} else {
				for (i = 0, l = swagger.tags.length; i < l; i++) {
					tag = swagger.tags[i];
					resources.push(tag);
					map[tag.name] = i;
				}
			}
		}

		/**
		 * parse operations
		 */
		function parseOperations(swagger, resources, form, map, defaultContentType, openPath) {
			var i,
				path,
				pathObject,
				pathParameters,
				httpMethod,
				operation,
				tag,
				resource,
				openModel;

			for (path in swagger.paths) {
				pathObject = swagger.paths[path];
				pathParameters = pathObject.parameters || [];
				delete pathObject.parameters;
				for (httpMethod in pathObject) {
					operation = pathObject[httpMethod];
					operation.description = trustHtml(operation.description);
					operation.produces = operation.produces || swagger.produces;
					form[operationId] = {
						responseType: operation.produces && operation.produces[0] || defaultContentType
					};
					operation.httpMethod = httpMethod;
					operation.path = path;
					operation.security = operation.security || swagger.security;
					openModel = openPath && openPath.match(new RegExp(operation.operationId + '-model-.*'));
					openModel = openModel && openModel[0];
					parseParameters(swagger, operation, pathParameters, form, defaultContentType, openModel);
					parseResponses(swagger, operation, openModel);
					operation.tags = operation.tags || ['default'];
					// map operation to resources
					for (i = 0; i < operation.tags.length; i++) {
						tag = operation.tags[i];
						if (typeof map[tag] === 'undefined') {
							map[tag] = resources.length;
							resources.push({
								name: tag
							});
						}
						resource = resources[map[tag]];
						resource.operations = resource.operations || [];
						operation.id = operationId++;
						operation.open = openPath && (openPath.match(new RegExp(operation.operationId + '$|' + resource.name + '\\*$')) || openModel);
						resource.operations.push(angular.copy(operation));
						if (operation.open) {
							resource.open = true;
						}
					}
				}
			}
		}

		/**
		 * compute path and operation parameters
		 */
		function computeParameters(swagger, pathParameters, operation) {
			var i, j, k, l,
				operationParameters = operation.parameters || [],
				parameters = [].concat(operationParameters),
				found,
				pathParameter,
				operationParameter;

			for (i = 0, l = pathParameters.length; i < l; i++) {
				found = false;
				pathParameter = swaggerModel.resolveReference(swagger, pathParameters[i]);

				for (j = 0, k = operationParameters.length; j < k; j++) {
					operationParameter = swaggerModel.resolveReference(swagger, operationParameters[j]);
					if (pathParameter.name === operationParameter.name && pathParameter.in === operationParameter.in) {
						// overriden parameter
						found = true;
						break;
					}
				}
				if (!found) {
					// add path parameter to operation ones
					parameters.push(pathParameter);
				}
			}
			return parameters;
		}

		/**
		 * parse operation parameters
		 */
		function parseParameters(swagger, operation, pathParameters, form, defaultContentType, openModel) {
			var i, l,
				param,
				model,
				parameters = operation.parameters = computeParameters(swagger, pathParameters, operation);

			for (i = 0, l = parameters.length; i < l; i++) {
				//TODO manage 'collectionFormat' (csv, multi etc.) ?
				//TODO manage constraints (pattern, min, max etc.) ?
				param = parameters[i] = swaggerModel.resolveReference(swagger, parameters[i]);
				param.id = paramId;
				param.type = swaggerModel.getType(param);
				param.description = trustHtml(param.description);
				if (param.items && param.items.enum) {
					param.enum = param.items.enum;
					param.default = param.items.default;
				}
				param.subtype = param.enum ? 'enum' : param.type;
				// put param into form scope
				form[operationId][param.name] = param.default || '';
				if (param.schema) {
					param.schema.json = swaggerModel.generateSampleJson(swagger, param.schema);
					model = swaggerModel.generateModel(swagger, param.schema, operation.operationId);
					param.schema.model = $sce.trustAsHtml(model);
					param.schema.display = openModel && model.match(new RegExp(openModel)) ? 0 : 1;
				}
				// fix consumes
				if (param.in === 'body') {
					operation.consumes = operation.consumes || swagger.consumes || [defaultContentType];
					form[operationId].contentType = operation.consumes && operation.consumes[0];
				} else if (param.in === 'formData') {
					operation.consumes = operation.consumes || [param.subtype === 'file' ? 'multipart/form-data' : 'application/x-www-form-urlencoded'];
					form[operationId].contentType = operation.consumes && operation.consumes[0];
				}
				paramId++;
			}
		}

		/**
		 * parse operation responses
		 */
		function parseResponses(swagger, operation, openModel) {
			var code,
				model,
				response,
				sampleJson;

			if (operation.responses) {
				for (code in operation.responses) {
					response = operation.responses[code] = swaggerModel.resolveReference(swagger, operation.responses[code]);
					response.description = trustHtml(response.description);
					if (response.schema) {
						if (response.examples && response.examples[operation.produces[0]]) {
							sampleJson = angular.toJson(response.examples[operation.produces[0]], true);
						} else {
							sampleJson = swaggerModel.generateSampleJson(swagger, response.schema);
						}
						response.schema.json = sampleJson;
						model = swaggerModel.generateModel(swagger, response.schema, operation.operationId);
						response.display = openModel && model.match(new RegExp(openModel)) ? 0 : 1;
						response.schema.model = $sce.trustAsHtml(model);
						if (code === '200' || code === '201') {
							operation.responseClass = response;
							operation.responseClass.display = response.display;
							operation.responseClass.status = code;
							parseHeaders(swagger, operation, response);
							delete operation.responses[code];
						} else {
							operation.hasResponses = true;
						}
					} else {
						operation.hasResponses = true;
					}
				}
			}
		}

		/**
		 * parse operation response headers
		 */
		function parseHeaders(swagger, operation, response) {
			if (response.headers) {
				operation.headers = response.headers;
				for (var name in operation.headers) {
					var header = operation.headers[name];
					header.type = swaggerModel.getType(header);
					if (header.type === 'array') {
						header.type = 'Array[' + swaggerModel.getType(header.items) + ']';
					}
					header.description = trustHtml(header.description);
				}
				delete response.headers;
			}
		}

		function cleanUp(resources, openPath, sortResources) {
			var i,
				resource,
				operations;

			for (i = 0; i < resources.length; i++) {
				resource = resources[i];
				operations = resources[i].operations;
				resource.open = resource.open || openPath === resource.name || openPath === resource.name + '*';
				if (!operations || (operations && operations.length === 0)) {
					resources.splice(i--, 1);
				}
			}
			if (sortResources) {
				// sort resources alphabeticaly
				resources.sort(function(a, b) {
					if (a.name > b.name) {
						return 1;
					} else if (a.name < b.name) {
						return -1;
					}
					return 0;
				});
			}
			swaggerModel.clearCache();
		}

		function trustHtml(text) {
			var trusted = text;
			if (typeof text === 'string' && trustedSources) {
				trusted = $sce.trustAsHtml(escapeChars(text));
			}
			// else ngSanitize MUST be added to app
			return trusted;
		}

		function escapeChars(text) {
			return text
				.replace(/&/g, '&amp;')
				.replace(/<([^\/a-zA-Z])/g, '&lt;$1')
				.replace(/"/g, '&quot;')
				.replace(/'/g, '&#039;');
		}

	}])
	.run(["swaggerModules", "swaggerParser", function(swaggerModules, swaggerParser) {
		swaggerModules.add(swaggerModules.PARSE, swaggerParser);
	}]);
/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.provider('swaggerTranslator', function() {

		var currentLang = 'en',
			allTranslations = {};

		/**
		 * set startup language
		 */
		this.setLanguage = function(lang) {
			currentLang = lang;
			return this;
		};

		/**
		 * add translations for a specific language code
		 */
		this.addTranslations = function(lang, translations) {
			var map = allTranslations[lang] = allTranslations[lang] || {};
			angular.merge(map, translations);
			return this;
		};

		this.$get = ["$rootScope", "$interpolate", function($rootScope, $interpolate) {
			return {
				/**
				 * change current used language
				 */
				useLanguage: function(lang) {
					if (typeof allTranslations[lang] === 'undefined') {
						console.error('AngularSwaggerUI: No translations found for language '+lang);
					}
					currentLang = lang;
					$rootScope.$emit('swaggerTranslateLangChanged');
				},
				/**
				 * get a localized message
				 */
				translate: function(key, values) {
					if (currentLang && allTranslations && allTranslations[currentLang] && allTranslations[currentLang][key]) {
						return $interpolate(allTranslations[currentLang][key])(values);
					} else {
						return key;
					}
				},
				/**
				 * get current used language
				 */
				language: function() {
					return currentLang;
				}
			};
		}];
	})
	.directive('swaggerTranslate', ["$rootScope", "$parse", "swaggerTranslator", function($rootScope, $parse, swaggerTranslator) {

		return {
			restrict: 'A',
			link: function(scope, element, attributes) {
				function translate() {
					var params;
					if (attributes.swaggerTranslateValue) {
						params = $parse(attributes.swaggerTranslateValue)(scope.$parent);
					}
					element.text(swaggerTranslator.translate(attributes.swaggerTranslate, params));
				}
				var unbind = $rootScope.$on('swaggerTranslateLangChanged', function() {
					translate();
				});
				scope.$on('$destroy', unbind);
				translate();
			}
		};

	}])
	.filter('swaggerTranslate', ["swaggerTranslator", function(swaggerTranslator) {

		var filter = function(input, option) {
			return swaggerTranslator.translate(input, option);
		};
		filter.$stateful = true;
		return filter;

	}]);
/*
 * Orange angular-swagger-ui - v0.4.4
 *
 * (C) 2015 Orange, all right reserved
 * MIT Licensed
 */
'use strict';

angular
	.module('swaggerUi')
	.config(["swaggerTranslatorProvider", function(swaggerTranslatorProvider) {

		swaggerTranslatorProvider
			.addTranslations('en', {
				infoContactCreatedBy: 'Created by {{name}}',
				infoContactUrl: 'See more at',
				infoContactEmail: 'Contact the developer',
				infoLicense: 'License: ',
				infoBaseUrl: 'BASE URL',
				infoApiVersion: 'API VERSION',
				infoHost: 'HOST',
				endPointToggleOperations: 'Open/Hide',
				endPointListOperations: 'List operations',
				endPointExpandOperations: 'Expand operations',
				operationDeprected: 'Warning: Deprecated',
				operationImplementationNotes: 'Implementation notes',
				externalDocs: 'External docs',
				headers: 'Response headers',
				headerName: 'Header',
				headerDescription: 'Description',
				headerType: 'Type',
				parameters: 'Parameters',
				parameterName: 'Parameter',
				parameterValue: 'Value',
				parameterDescription: 'Description',
				parameterType: 'Parameter type',
				parameterDataType: 'Data type',
				parameterOr: ' or ',
				parameterRequired: '(required)',
				parameterModel: 'Model',
				parameterSchema: 'Example value',
				parameterContentType: 'Parameter content type',
				parameterDefault: '{{default}} (default)',
				parameterSetValue: 'Click to set as parameter value',
				responseClass: 'Response class (status {{status}})',
				responseModel: 'Model',
				responseSchema: 'Example value',
				responseContentType: 'Response content type',
				responses: 'Response messages',
				responseCode: 'HTTP status code',
				responseReason: 'Reason',
				responseHide: 'Hide response',
				modelOptional: 'optional',
				modelOr: ' or ',
				explorerUrl: 'Request URL',
				explorerBody: 'Response body',
				explorerCode: 'Response code',
				explorerHeaders: 'Response headers',
				explorerLoading: 'Loading...',
				explorerTryIt: 'Try it out!',
				errorNoParserFound: 'No parser found for Swagger specification of type {{type}} and version {{version}}',
				errorParseFailed: 'Failed to parse Swagger specification: {{message}}',
				errorJsonParse: 'Failed to parse JSON',
				errorNoYamlParser: 'No YAML parser found, please make sure to include js-yaml library',
				authRequired: 'Authorization required',
				authAvailable: 'Available authorizations',
				apiKey: 'API key authorization',
				authParamName: 'Name',
				authParamType: 'In',
				authParamValue: 'Value',
				basic: 'Basic authorization',
				authLogin: 'Login',
				authPassword: 'Password',
				oauth2: 'oAuth2 authorization',
				authOAuthDesc: 'Scopes are used to grant an application different levels of access to data on behalf of the end user. Each API may declare one or more scopes. API requires the following scopes. Select which ones you want to grant.',
				authAuthorizationUrl: 'Authorization URL',
				authFlow: 'Flow',
				authTokenUrl: 'Token URL',
				authScopes: 'Scopes',
				authCancel: 'Cancel',
				authAuthorize: 'Authorize'
			});

	}]);
angular.module('swaggerUi').run(['$templateCache', function($templateCache) {
  $templateCache.put('templates/endpoint.html',
    '<div id="{{api.name}}" class="clearfix" ng-class="api.css"> <ul id="{{api.name}}*" class="list-inline pull-left endpoint-heading"> <li> <h4> <a ng-click="api.open=!api.open;permalink(api.name)" ng-bind="api.name"></a> <span ng-if="api.description"> : <span ng-bind="api.description"></span></span> </h4> </li> </ul> <ul class="list-inline pull-right endpoint-actions"> <li> <a ng-click="api.open=!api.open;permalink(api.name)" swagger-translate="endPointToggleOperations"></a> </li> <li> <a ng-click="expand(api);permalink(api.name)" swagger-translate="endPointListOperations"></a> </li> <li> <a ng-click="expand(api,true);permalink(api.name+\'*\')" swagger-translate="endPointExpandOperations"></a> </li> </ul> </div> <ul class="list-unstyled operations" ng-class="api.css" ng-if="api.open"> <li ng-repeat="op in api.operations track by $index" class="operation" ng-class="[op.httpMethod,op.css]" ng-include="\'templates/operation.html\'"></li> </ul>');
  $templateCache.put('templates/operation.html',
    '<div id="{{op.operationId}}" class="heading"> <a ng-click="op.open=!op.open;permalink(op.operationId)"> <div class="clearfix"> <span class="http-method text-uppercase" ng-bind="op.httpMethod"></span> <span class="path" ng-class="{deprecated:op.deprecated}" ng-bind="op.path"></span> <span class="description pull-right" ng-bind="op.summary"></span> </div> </a> </div> <div class="content" ng-if="op.open"> <div class="h5" ng-if="op.deprecated" swagger-translate="operationDeprected"></div> <div ng-if="op.description"> <h5 swagger-translate="operationImplementationNotes"></h5> <p ng-bind-html="op.description"></p> </div> <div ng-if="op.externalDocs"> <h5 swagger-translate="externalDocs"></h5> <span ng-bind-html="op.externalDocs.description"></span> <a target="_blank" ng-href="{{op.externalDocs.url}}" ng-bind="op.externalDocs.url"></a> </div> <button ng-if="op.security.length>0" ng-click="auth(op)" class="auth-required" ng-class="{valid:authValid(op)}" title="{{\'authRequired\'|swaggerTranslate}}">!</button> <form role="form" name="explorerForm" ng-submit="explorerForm.$valid&&submitExplorer(op)"> <div ng-if="op.responseClass" class="response"> <h5 swagger-translate="responseClass" swagger-translate-value="op.responseClass"></h5> <div ng-if="op.responseClass.display!=-1"> <ul class="list-inline schema"> <li><a ng-click="op.responseClass.display=0" ng-class="{active:op.responseClass.display==0}" swagger-translate="responseModel"></a></li> <li><a ng-click="op.responseClass.display=1" ng-class="{active:op.responseClass.display==1}" swagger-translate="responseSchema"></a></li> </ul> <pre class="model" ng-if="op.responseClass.display==0" ng-bind-html="op.responseClass.schema.model"></pre> <pre class="model-schema" ng-if="op.responseClass.display==1" ng-bind="op.responseClass.schema.json"></pre> </div> </div> <div ng-if="op.headers" class="table-responsive"> <h5 swagger-translate="headers"></h5> <table class="table table-condensed headers"> <thead> <tr> <th class="name" swagger-translate="headerName"> <th class="desc" swagger-translate="headerDescription"> <th class="type" swagger-translate="headerType">   <tbody> <tr ng-repeat="(name,header) in op.headers track by $index" ng-class="header.css" class="response-header"> <td class="bold" ng-bind="name"> <td ng-bind-html="header.description"> <td ng-bind="header.type">   </table> </div> <div ng-if="op.parameters&&op.parameters.length>0" class="table-responsive"> <h5 swagger-translate="parameters"></h5> <table class="table table-condensed parameters"> <thead> <tr> <th class="name" swagger-translate="parameterName"> <th class="value" swagger-translate="parameterValue"> <th class="desc" swagger-translate="parameterDescription"> <th class="type" swagger-translate="parameterType"> <th class="data" swagger-translate="parameterDataType">   <tbody> <tr ng-repeat="param in op.parameters track by $index" ng-include="\'templates/parameter.html\'" class="operation-parameter" ng-class="[param.css,\'operation-parameter-\'+param.in]">  </table> </div> <div ng-if="op.produces" class="content-type"> <label for="responseContentType{{op.id}}" swagger-translate="responseContentType"></label> <select ng-model="form[op.id].responseType" ng-options="item for item in op.produces track by item" id="responseContentType{{op.id}}" name="responseContentType{{op.id}}" required></select> </div> <div class="table-responsive" ng-if="op.hasResponses"> <h5 swagger-translate="responses"></h5> <table class="table responses"> <thead> <tr> <th class="code" swagger-translate="responseCode"> <th swagger-translate="responseReason"> <th swagger-translate="responseModel">   <tbody> <tr ng-repeat="(code,resp) in op.responses track by $index" ng-include="\'templates/response.html\'" class="operation-response" ng-class="resp.css">  </table> </div> <div ng-if="apiExplorer"> <button class="btn btn-default" ng-click="op.explorerResult=false;op.hideExplorerResult=false" type="submit" ng-disabled="op.loading" ng-bind="op.loading?\'explorerLoading\':\'explorerTryIt\'|swaggerTranslate"></button> <a class="hide-try-it" ng-if="op.explorerResult&&!op.hideExplorerResult" ng-click="op.hideExplorerResult=true" swagger-translate="responseHide"></a> </div> </form> <div ng-if="op.explorerResult" ng-show="!op.hideExplorerResult" class="explorer-result"> <h5 swagger-translate="explorerUrl"></h5> <pre ng-bind="op.explorerResult.url" class="explorer-url"></pre> <h5 swagger-translate="explorerBody"></h5> <pre ng-bind="op.explorerResult.response.body" class="explorer-body"></pre> <h5 swagger-translate="explorerCode"></h5> <pre ng-bind="op.explorerResult.response.status" class="explorer-status"></pre> <h5 swagger-translate="explorerHeaders"></h5> <pre ng-bind="op.explorerResult.response.headers" class="explorer-headers"></pre> </div> </div>');
  $templateCache.put('templates/parameter.html',
    '<td ng-class="{bold:param.required}" class="operation-parameter-name"> <label for="param{{param.id}}" ng-bind="param.name"></label>  <td ng-class="{bold:param.required}" class="operation-parameter-value"> <div ng-if="apiExplorer"> <div ng-if="param.in!=\'body\'" ng-switch="param.subtype"> <input ng-switch-when="file" type="file" file-input ng-model="form[op.id][param.name]" id="param{{param.id}}" placeholder="{{param.required?\'parameterRequired\':\'\'|swaggerTranslate}}" ng-required="param.required" name="{{param.name}}"> <select ng-switch-when="enum" ng-model="form[op.id][param.name]" id="param{{param.id}}" name="{{param.name}}"> <option ng-repeat="value in param.enum" value="{{value}}" ng-bind="value+(param.default==value?\'parameterDefault\':\'\'|swaggerTranslate)" ng-selected="param.default==value"> </select> <input ng-switch-default type="text" ng-model="form[op.id][param.name]" id="param{{param.id}}" placeholder="{{param.required?\'parameterRequired\':\'\'|swaggerTranslate}}" ng-required="param.required" name="{{param.name}}"> </div> <div ng-if="param.in==\'body\'"> <textarea id="param{{param.id}}" ng-model="form[op.id][param.name]" ng-required="param.required" name="{{param.name}}"></textarea> <br> <div ng-if="op.consumes" class="content-type"> <label for="bodyContentType{{op.id}}" swagger-translate="parameterContentType"></label> <select ng-model="form[op.id].contentType" id="bodyContentType{{op.id}}" name="bodyContentType{{op.id}}" ng-options="item for item in op.consumes track by item"></select> </div> </div> </div> <div ng-if="!apiExplorer"> <div ng-if="param.in!=\'body\'"> <div ng-if="param.default" swagger-translate="parameterDefault" swagger-translate-value="param.default"></div> <div ng-if="param.enum"> <span ng-repeat="value in param.enum track by $index">{{value}}<span ng-if="!$last" swagger-translate="parameterOr"></span></span> </div> <div ng-if="param.required"><strong swagger-translate="parameterRequired"></strong></div> </div> </div>  <td ng-class="{bold:param.required}" ng-bind-html="param.description" class="operation-parameter-description"> <td ng-bind="param.in" class="operation-parameter-in"> <td ng-if="param.type" ng-switch="param.type" class="operation-parameter-type"> <span ng-switch-when="array" ng-bind="\'Array[\'+param.items.type+\']\'"></span> <span ng-switch-default ng-bind="param.type"></span>  <td ng-if="param.schema" class="operation-parameter-model"> <ul class="list-inline schema"> <li><a ng-click="param.schema.display=0" ng-class="{active:param.schema.display==0}" swagger-translate="parameterModel"></a></li> <li><a ng-click="param.schema.display=1" ng-class="{active:param.schema.display==1}" swagger-translate="parameterSchema"></a></li> </ul> <pre class="model" ng-if="param.schema.display==0&&param.schema.model" ng-bind-html="param.schema.model"></pre> <div class="model-schema" ng-if="param.schema.display==1&&param.schema.json"> <pre ng-bind="param.schema.json" ng-click="form[op.id][param.name]=param.schema.json" aria-described-by="help-{{param.id}}"></pre> <div id="help-{{param.id}}" swagger-translate="parameterSetValue"></div> </div> ');
  $templateCache.put('templates/response.html',
    '<td ng-bind="code" class="operation-response-code"> <td ng-bind-html="resp.description" class="operation-response-description"> <td class="operation-response-model"> <ul ng-if="resp.schema&&resp.schema.model&&resp.schema.json" class="list-inline schema"> <li><a ng-click="resp.display=0" ng-class="{active:resp.display==0}" swagger-translate="responseModel"></a></li> <li><a ng-click="resp.display=1" ng-class="{active:resp.display==1}" swagger-translate="responseSchema"></a></li> </ul> <pre class="model" ng-if="resp.display==0&&resp.schema&&resp.schema.model" ng-bind-html="resp.schema.model"></pre> <pre class="model-schema" ng-if="resp.display==1&&resp.schema&&resp.schema.json" ng-bind="resp.schema.json"></pre> ');
  $templateCache.put('templates/swagger-ui.html',
    '<div class="swagger-ui" aria-live="polite" aria-relevant="additions removals"> <div class="api-name"> <h3 ng-bind="infos.title"></h3> </div> <div class="api-description" ng-bind-html="infos.description"></div> <div class="external-docs" ng-if="infos.externalDocs"> <span ng-bind-html="infos.externalDocs.description"></span> <a target="_blank" ng-href="{{infos.externalDocs.url}}">{{infos.externalDocs.url}}</a> </div> <div class="api-infos"> <div class="api-infos-contact" ng-if="infos.contact"> <div ng-if="infos.contact.name" class="api-infos-contact-name"><span swagger-translate="infoContactCreatedBy" swagger-translate-value="infos.contact"></span></div> <div ng-if="infos.contact.url" class="api-infos-contact-url"><span swagger-translate="infoContactUrl"></span> <a ng-href="{{infos.contact.url}}" ng-bind="infos.contact.url"></a></div> <a ng-if="infos.contact.email" class="api-infos-contact-url" ng-href="mailto:{{infos.contact.email}}?subject={{infos.title}}" swagger-translate="infoContactEmail"></a> </div> <div class="api-infos-license" ng-if="infos.license"> <span swagger-translate="infoLicense"></span><a ng-href="{{infos.license.url}}" ng-bind="infos.license.name"></a> </div> </div> <ul class="list-unstyled endpoints"> <li ng-repeat="api in resources track by $index" class="endpoint" ng-class="{active:api.open}" ng-include="\'templates/endpoint.html\'"></li> </ul> <div class="api-version clearfix" ng-if="infos"> [<span swagger-translate="infoBaseUrl"></span>: <span class="h4" ng-bind="infos.basePath"></span>, <span swagger-translate="infoApiVersion"></span>: <span class="h4" ng-bind="infos.version"></span>, <span swagger-translate="infoHost"></span>: <span class="h4" ng-bind="infos.scheme"></span>://<span class="h4" ng-bind="infos.host"></span>] <a ng-if="validatorUrl!=\'false\'" target="_blank" ng-href="{{validatorUrl}}/debug?url={{url}}"><img class="pull-right swagger-validator" ng-src="{{validatorUrl}}?url={{url}}"></a> </div> </div>');
}]);
