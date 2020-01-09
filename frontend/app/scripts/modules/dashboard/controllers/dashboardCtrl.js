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
angular.module('serviceCenter.dashboard', [])
    .controller('dashboardController', ['$scope', '$state', 'apiConstant', 'httpService', '$q', '$interval', function($scope, $state, apiConstant, httpService, $q, $interval) {

        $scope.getAllServices = function() {
            $(".loader").show();
            $scope.dashboardInfo = [{
                    count: 0,
                    title: "services",
                    iconName: "fa fa-cog"
                },
                {
                    count: 0,
                    title: "instances",
                    iconName: "fa fa-cogs"
                },
                {
                    count: 0,
                    title: "providers",
                    iconName: "fa fa-arrow-circle-right"
                },
                {
                    count: 0,
                    title: "consumers",
                    iconName: "fa fa-arrow-circle-left"
                }
            ];
            $scope.instanceStat = [{
                    count: 0,
                    title: "starting",
                    percent: 0,
                    status: "STARTING"
                },
                {
                    count: 0,
                    title: "testing",
                    percent: 0,
                    status: "TESTING"
                },
                {
                    count: 0,
                    title: "up",
                    percent: 0,
                    status: "UP"
                },
                {
                    count: 0,
                    title: "outOfService",
                    percent: 0,
                    status: "OUTOFSERVICE"
                },
                {
                    count: 0,
                    title: "down",
                    percent: 0,
                    status: "DOWN"
                }
            ];

            $scope.runningServices = [];
            $scope.stoppedServices = [];
            $scope.startingServices = [];
            $scope.outOfServiceServices = [];
            $scope.testingServices = [];

            $scope.totalProviders = [];
            $scope.totalConsumers = [];

            $scope.services = [];
            $scope.instances = [];

            var url = apiConstant.api.allServices.url;
            var method = apiConstant.api.allServices.method;
            httpService.apiRequest(url, method, null, null).then(function(response) {
                $(".loader").hide();
                if (response && response.data && response.data.allServicesDetail) {
                    $scope.dashboardInfo[0].count = response.data.allServicesDetail.length;
                    response.data.allServicesDetail.forEach(function(services) {
                        $scope.services.push(services.microService);
                        if (services.instances) {
                            services.instances.forEach(function(instance) {
                                instance.serviceName = services.microService.serviceName;
                                instance.appId = services.microService.appId;
                                $scope.instances.push(instance);

                                if (instance.status == "STARTING") {
                                    $scope.startingServices.push(services);
                                }
                                if (instance.status == "UP") {
                                    $scope.runningServices.push(services);
                                }
                                if (instance.status == "DOWN") {
                                    $scope.stoppedServices.push(services);
                                }
                                if (instance.status == "OUTOFSERVICE") {
                                    $scope.outOfServiceServices.push(services);
                                }
                                if (instance.status == "TESTING") {
                                    $scope.testingServices.push(services);
                                }
                            });
                            $scope.dashboardInfo[1].count = $scope.dashboardInfo[1].count + services.instances.length;
                        }

                        if (services.providers) {
                            services.providers.forEach(function(provider) {
                                var isProvider = 0;
                                for (var p = 0; p < $scope.totalProviders.length; p++) {
                                    if (provider.serviceId == $scope.totalProviders[p]) {
                                        isProvider = 1;
                                    }
                                }
                                if (isProvider == 0) {
                                    $scope.totalProviders.push(provider.serviceId);
                                }
                            })
                        }

                        if (services.consumers) {
                            services.consumers.forEach(function(consumer) {
                                var isConsumer = 0;
                                for (var c = 0; c < $scope.totalConsumers.length; c++) {
                                    if (consumer.serviceId == $scope.totalConsumers[c]) {
                                        isConsumer = 1;
                                    }
                                }
                                if (isConsumer == 0) {
                                    $scope.totalConsumers.push(consumer.serviceId);
                                }
                            })
                        }
                    });

                    $scope.instanceStat[0].percent = Math.round($scope.startingServices.length / $scope.dashboardInfo[1].count) * 100;
                    $scope.instanceStat[1].percent = Math.round($scope.testingServices.length / $scope.dashboardInfo[1].count) * 100;
                    $scope.instanceStat[2].percent = Math.round($scope.runningServices.length / $scope.dashboardInfo[1].count) * 100;
                    $scope.instanceStat[3].percent = Math.round($scope.outOfServiceServices.length / $scope.dashboardInfo[1].count) * 100;
                    $scope.instanceStat[4].percent = Math.round($scope.stoppedServices.length / $scope.dashboardInfo[1].count) * 100;

                    $scope.instanceStat[0].count = $scope.startingServices.length;
                    $scope.instanceStat[1].count = $scope.testingServices.length;
                    $scope.instanceStat[2].count = $scope.runningServices.length;
                    $scope.instanceStat[3].count = $scope.outOfServiceServices.length;
                    $scope.instanceStat[4].count = $scope.stoppedServices.length;

                    $scope.dashboardInfo[2].count = $scope.totalProviders.length;
                    $scope.dashboardInfo[3].count = $scope.totalConsumers.length;
                }
            }, function(error) {
                $(".loader").hide();
            });
        };

        $scope.goToServices = function(status) {
            $state.go("sc.instances", {
                status: status
            });
        };

        $scope.getAllServices();

        var dashboardRefresh = $interval(function(){
            $scope.getAllServices();
        }, 10000);
 
        $scope.$on('$destroy', function(){
            $interval.cancel(dashboardRefresh);
        })
    }]);