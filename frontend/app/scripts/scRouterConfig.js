/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
'use strict';
angular.module('serviceCenter.router', [])
    .config(['$stateProvider', '$urlRouterProvider', function($stateProvider, $urlRouterProvider) {
        $urlRouterProvider.otherwise('/sc/dashboard');
        $stateProvider
            .state('sc', {
                url: '/sc',
                abstract: true,
                templateUrl: 'scripts/views/index.html',
                controller: 'serviceCenterController'
            })
            .state('sc.dashboard', {
                url: '/dashboard',
                views: {
                    'base': {
                        templateUrl: 'scripts/modules/dashboard/views/dashboard.html',
                        controller: 'dashboardController'
                    }
                }
            })
            .state('sc.allServices', {
                url: '/services/:status',
                views: {
                    'base': {
                        templateUrl: 'scripts/modules/serviceCenter/views/servicesList.html',
                        controller: 'servicesListController'
                    }
                }
            })
            .state('sc.info', {
                url: '/:serviceId',
                abstract: true,
                views: {
                    'base': {
                        templateUrl: 'scripts/modules/serviceCenter/views/serviceInfo.html',
                        controller: 'serviceInfoController'
                    }
                },
                resolve: {
                    serviceInfo: ['$q', 'httpService', 'commonService', 'apiConstant', '$stateParams', '$state', function($q, httpService, commonService, apiConstant, $stateParams, $state) {
                        $(".loader").show();
                        var serviceId = $stateParams.serviceId;
                        var deferred = $q.defer();
                        var url = apiConstant.api.particularService.url.replace('{{serviceId}}', serviceId);
                        var method = apiConstant.api.particularService.method;
                        httpService.apiRequest(url, method, null, null, null).then(function(response) {
                            $(".loader").hide();
                            deferred.resolve(response);
                        }, function(error) {
                            $(".loader").hide();
                            deferred.reject(error);
                            $state.go("sc.dashboard");
                        });
                        return deferred.promise;
                    }]
                }
            })
            .state('sc.info.instance', {
                url: '/instance',
                views: {
                    "info": {
                        templateUrl: 'scripts/modules/serviceCenter/views/serviceInstance.html'
                    }
                }
            })
            .state('sc.info.provider', {
                url: '/provider',
                views: {
                    "info": {
                        templateUrl: 'scripts/modules/serviceCenter/views/serviceProvider.html'
                    }
                }
            })
            .state('sc.info.consumer', {
                url: '/consumer',
                views: {
                    "info": {
                        templateUrl: 'scripts/modules/serviceCenter/views/serviceConsumer.html'
                    }
                }
            })
            .state('sc.info.properties', {
                url: '/properties',
                views: {
                    "info": {
                        templateUrl: 'scripts/modules/serviceCenter/views/serviceProperties.html'
                    }
                }
            })
            .state('sc.info.schema', {
                url: '/schema',
                views: {
                    "info": {
                        templateUrl: 'scripts/modules/serviceCenter/views/schema.html',
                        controller: 'schemaController'
                    }
                },
                resolve: {
                    servicesInfo: ['$q', 'httpService', 'apiConstant', '$stateParams', function($q, httpService,  apiConstant, $stateParams) {
                        $(".loader").show();
                        var deferred = $q.defer();
                        var serviceId = $stateParams.serviceId;
                        var url = apiConstant.api.particularService.url.replace('{{serviceId}}', serviceId);
                        var method = apiConstant.api.particularService.method;
                        httpService.apiRequest(url, method, null, null, null).then(function(response) {
                            $(".loader").hide();
                            deferred.resolve(response);
                        }, function(error) {
                            $(".loader").hide();
                            deferred.reject(error);
                        });
                        return deferred.promise;
                    }]
                }
            })
            .state('sc.instances', {
                url: '/instances',
                views: {
                    'base': {
                        templateUrl: 'scripts/modules/instances/views/instances.html',
                        controller: 'instancesController'
                    }
                }
            })
            .state('sc.topology', {
                url: '/topology',
                views: {
                    'base': {
                        templateUrl: 'scripts/modules/topology/views/topology.html',
                        controller: 'topologyController'
                    }
                }
            })
    }]);