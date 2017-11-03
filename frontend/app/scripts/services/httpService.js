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
angular.module('serviceCenter')
	.service('httpService', ['$http', '$q', '$mdDialog', 'apiConstant', function($http, $q, $mdDialog, apiConstant){

		function apiRequest(requrl, method, payload, headers, nopopup){
            var tenant = localStorage.getItem('tenant');
            if(!tenant || tenant == undefined || tenant == null){
                tenant = 'default';
                localStorage.setItem('tenant', 'default');
            }

			var defer = $q.defer();
            if(undefined === requrl || null === requrl || undefined === method || null === method){
                defer.reject("invalid params");
                return defer.promise;
            }
            var baseUrl = apiConstant.endPoint.ip + ':' + apiConstant.endPoint.port;
            if(undefined === headers || null === headers){
                headers = {
                    'x-domain-name' : tenant
                };
            }else{
                headers['x-domain-name'] = tenant;
            }

            var url = baseUrl + '/'+ requrl;
            $http({
                url: url,
                method: method,
                data: payload,
                headers : headers
            }).then(function(response) {
                defer.resolve(response);
            }, function(error) {
                if(nopopup){
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
                    skipHide : true,
                    controller: function($scope, $mdDialog, error) {
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
			apiRequest : apiRequest
		};
	}]);
