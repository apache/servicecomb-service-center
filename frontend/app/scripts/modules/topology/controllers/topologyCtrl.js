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
angular.module('serviceCenter.topology', [])
    .controller('topologyController', ['$scope', 'httpService', 'apiConstant', function($scope, httpService, apiConstant) {

        $scope.getServices = function(appId) {
            $(".loader").show();
            $scope.allAppId = [];
            $scope.microServices = [];
            $scope.prosAndCons = [];
            $scope.appId = '';

            var url = apiConstant.api.allServices.url;
            var method = apiConstant.api.allServices.method;
            httpService.apiRequest(url, method).then(function(response) {
                $(".loader").hide();
                if (response && response.data && response.data.allServicesDetail && response.data.allServicesDetail.length > 0) {
                    $scope.allAppId.push("All");
                    $scope.appId = appId ? appId : "All";
                    $scope.allServicesDetail = response.data.allServicesDetail;

                    angular.forEach(response.data.allServicesDetail, function(service) {
                        if (!$scope.allAppId.includes(service.microService.appId) && service.microService.serviceName.toLowerCase() !== 'servicecenter') {
                            $scope.allAppId.push(service.microService.appId);
                        }
                    });

                    if ($scope.appId === "All") {
                        $scope.microServices = [];
                        $scope.prosAndCons = [];
                        angular.forEach($scope.allAppId, function(appId) {
                            $scope.getDependencies(appId);
                        });
                        createTopology();
                    } else {
                        $scope.microServices = [];
                        $scope.prosAndCons = [];
                        $scope.getDependencies(appId);
                        createTopology();
                    }
                } else {
                    $scope.allAppId = [];
                    document.getElementById("topology").innerHTML = "";
                }
            }, function(error) {
                $(".loader").hide();
                $scope.allAppId = [];
                document.getElementById("topology").innerHTML = "";
            });
        }
        $scope.getServices();

        $scope.getDependencies = function(appId) {
            angular.forEach($scope.allServicesDetail, function(service) {
                if (appId === service.microService.appId && service.microService.serviceName.toLowerCase() !== 'servicecenter') {
                    var objMic = {};
                    var providers = [];
                    var consumers = [];
                    var showStatus = false;
                    var version, registerBy;

                    if (service.providers) {
                        angular.forEach(service.providers, function(provider) {
                            var objPro = {};
                            objPro.from = service.microService.serviceName + service.microService.appId + "#" + service.microService.version;
                            objPro.to = provider.serviceName + provider.appId + "#" + provider.version;
                            $scope.prosAndCons.push(objPro);
                            providers.push(provider.serviceName + "#" + provider.version);
                        })
                    }

                    if (service.consumers) {
                        angular.forEach(service.consumers, function(consumer) {
                            consumers.push(consumer.serviceName + "#" + consumer.version)
                        })
                    }

                    if (service.instances) {
                        angular.forEach(service.instances, function(instance) {
                            if (instance.status.toLowerCase() === 'up') {
                                showStatus = true;
                            }
                        })
                    }

                    objMic.id = service.microService.serviceName + service.microService.appId + "#" + service.microService.version;
                    objMic.label = service.microService.serviceName + "#" + service.microService.version;

                    providers = providers.length > 0 ? "<br> Providers : " + providers : "";
                    consumers = consumers.length > 0 ? "<br> Consumers : " + consumers : "";
                    status = showStatus ? "<br> Status : up" : "";
                    version = "<br> Version : " + service.microService.version;
                    registerBy = "<br> Register By :" + service.microService.registerBy;
                    objMic.title = service.microService.serviceName + status + version + registerBy + providers + consumers;
                    $scope.microServices.push(objMic);
                }
            });

        }

        function createTopology() {
            var microServices = new vis.DataSet($scope.microServices);
            var prosAndCons = new vis.DataSet($scope.prosAndCons);
            var container = document.getElementById('topology');
            var data = {
                nodes: microServices,
                edges: prosAndCons
            };

            var options = {
                autoResize: false,
                height: (screen.height - 300).toString(),
                width: (screen.width - 325).toString(),
                physics: {
                    enabled: true
                },
                edges: {
                    arrows: {
                        to: {
                            enabled: true,
                            scaleFactor: 1,
                            type: 'arrow'
                        }
                    },
                    arrowStrikethrough: false,
                    length: 170
                },
                nodes: {
                    shape: "circle",
                    color: {
                        border: "rgba(59,50,233,1)",
                        background: "rgb(255,255,255)",
                        highlight: {
                            border: "rgba(226,77,233,1)",
                            background: "rgba(130,38,255,1)"
                        },
                        hover: {
                            border: "rgba(226,77,233,1)",
                            background: "rgba(130,38,255,1)"
                        }
                    },
                    font: {
                        size: 17
                    },
                    shapeProperties: {
                        borderRadius: 5
                    }
                },
                interaction: {
                    navigationButtons: true,
                    hover: true
                }
            };
            var topology = new vis.Network(container, data, options);
        }

    }]);