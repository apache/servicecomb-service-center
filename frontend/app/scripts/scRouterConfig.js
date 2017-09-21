//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
'use strict';
angular.module('serviceCenter.router', [])
	.config(['$stateProvider', '$urlRouterProvider', function($stateProvider, $urlRouterProvider) {
    $urlRouterProvider.otherwise('/sc/services');
    $stateProvider
        .state('sc', {
            url: '/sc',
            abstract: true,
            templateUrl: 'scripts/views/index.html',
            controller: 'serviceCenterController as scCtrl'
        })
        .state('sc.allServices', {
            url: '/services',
            views:{
                'base' :{
                    templateUrl: 'scripts/serviceCenter/views/servicesList.html',
                    controller: 'servicesListController as services'
                }
            }
        })
        .state('sc.allInstances', {
            url: '/instances',
            views:{
                'base' :{
                    templateUrl: 'scripts/instances/views/instanceList.html',
                    controller: 'instancesListController as instances',
                }
            },
            resolve: {
                servicesList: ['$q', 'httpService', 'apiConstant',function($q, httpService, apiConstant){
                    var deferred = $q.defer();
                    var url = apiConstant.api.microservice.url;
                    var method = apiConstant.api.microservice.method;
                    httpService.apiRequest(url,method).then(function(response){
                        if(response && response.data && response.data.services){
                            deferred.resolve(response.data.services);
                        }
                        else {
                            deferred.resolve(response);
                        }
                    },function(error){
                        deferred.reject(error);
                    });
                    return deferred.promise;
                }]
            }
        })
        .state('sc.info',{
            url: '/:serviceId',
            abstract: true,
            views: {
                'base': {
                    templateUrl: 'scripts/serviceCenter/views/serviceInfo.html',
                    controller: 'serviceInfoController as serviceInfo'
                }
            },
            resolve: {
                serviceInfo: ['$q', 'httpService', 'commonService', 'apiConstant', '$stateParams', function($q, httpService, commonService, apiConstant, $stateParams){
                    $(".loader").show();
                    var serviceId = $stateParams.serviceId;
                    var deferred = $q.defer();
                    var url = apiConstant.api.microservice.url;
                    var method = apiConstant.api.microservice.method;
                    httpService.apiRequest(url,method).then(function(response){
                        $(".loader").hide();
                        if(response && response.data && response.data.services){
                            response.data.services.forEach(function(services){
                                if(services.serviceId == serviceId){
                                    var selectedService = {
                                        serviceName: services.serviceName.toUpperCase(),
                                        status: services.status.toLowerCase(),
                                        appId: services.appId.toLowerCase(),
                                        version: services.version,
                                        createdAt: commonService.timeFormat(services.timestamp),
                                        serviceId: services.serviceId
                                    };
                                    deferred.resolve(selectedService);
                                }else {
                                    deferred.resolve(response);
                                }
                            });
                        }
                        else {
                            deferred.resolve(response);
                        }
                    },function(error){
                        $(".loader").hide();
                        deferred.reject(error);
                    });
                    return deferred.promise;
                }]
            }
        })
        .state('sc.info.instance', {
            url: '/instance',
            views: {
                "info" : {
                    templateUrl: 'scripts/serviceCenter/views/serviceInstance.html'
                }
            }
        })
        .state('sc.info.provider', {
            url: '/provider',
            views: {
                "info" : {
                    templateUrl: 'scripts/serviceCenter/views/serviceProvider.html'
                }
            }
        })
        .state('sc.info.consumer', {
            url: '/consumer',
            views: {
                "info" : {
                    templateUrl: 'scripts/serviceCenter/views/serviceConsumer.html'
                }
            }
        })
        .state('sc.info.schema', {
            url: '/schema',
            views: {
                "info" : {
                    templateUrl: 'scripts/serviceCenter/views/schema.html'
                }
            }
        })
        .state('error', {
        	url: '/error',
        	templateUrl: 'views/error.html'
        });
}]);
