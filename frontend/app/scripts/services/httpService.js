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
angular.module('serviceCenter')
    .service('httpService', ['$http', '$q', '$mdDialog', 'apiConstant', function($http, $q, $mdDialog, apiConstant) {

        function apiRequest(requrl, method, payload, headers, nopopup) {
            var tenant = localStorage.getItem('tenant');
            if (!tenant || tenant == undefined || tenant == null) {
                tenant = 'default';
                localStorage.setItem('tenant', 'default');
            }

            var defer = $q.defer();
            if (undefined === requrl || null === requrl || undefined === method || null === method) {
                defer.reject("invalid params");
                return defer.promise;
            }
            var baseUrl = "/sc"
            if (undefined === headers || null === headers) {
                headers = {
                    'x-domain-name': tenant
                };
            } else {
                headers['x-domain-name'] = tenant;
            }

            var url = baseUrl + '/' + requrl;
            $http({
                url: url,
                method: method,
                data: payload,
                headers: headers
            }).then(function(response) {
                defer.resolve(response);
            }, function(error) {
                if (nopopup) {
                    defer.reject(error);
                    return;
                }
                var parentEl = angular.element(document.body);
                $mdDialog.show({
                    parent: parentEl,
                    templateUrl: 'views/serverError.html',
                    locals: {
                        error: error
                    },
                    skipHide: true,
                    controller: function($scope, $mdDialog, error) {
                        switch (error.status) {
                            case 502:
                                error.data = "Service-Center is not reachable"
                                break;
                            case 500:
                                error.data = "Mostly probably the Service-Center is not running or crashed"
                                break;
                            case 400:
                                error.data = "The request is in-appropriate"
                                break;
                            case 404:
                                error.data = "Requested entity was not found"
                                break;
                        }
                        $scope.error = error;
                        $scope.closeDialog = function() {
                            $mdDialog.hide();
                        };
                    }
                });
                defer.reject(error);
            });
            return defer.promise;
        }

        return {
            apiRequest: apiRequest
        };
    }]);
