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
angular.module('serviceCenter.sc')
	.controller('serviceInfoController', ['$scope', 'httpService', 'commonService', '$q', 'apiConstant', '$state',
	 '$stateParams','serviceInfo' ,function($scope, httpService, commonService, $q, apiConstant, $state, $stateParams, serviceInfo){

			var serviceId = $stateParams.serviceId;
			if(serviceInfo && serviceInfo.data && serviceInfo.data.services){
				serviceInfo.data.services.forEach(function(services){
	                if(services.serviceId == serviceId){
	                    $scope.serviceDetail = {
	                        serviceName: services.serviceName,
	                        status: services.status,
	                        appId: services.appId,
	                        version: services.version,
	                        createdAt: commonService.timeFormat(services.timestamp),
	                        serviceId: services.serviceId
	                    };
	                }
            	});
			}
			

			var apis = [];
			var instanceUrl = apiConstant.api.instances.url;
			var instanceApi = instanceUrl.replace('{{serviceId}}', serviceId);
			apis.push(instanceApi);
			var consumerUrl = apiConstant.api.consumer.url;
			var consumerApi = consumerUrl.replace('{{consumerId}}', serviceId);
			apis.push(consumerApi);
			var providerUrl = apiConstant.api.provider.url;
			var providerApi = providerUrl.replace('{{providerId}}', serviceId);
			apis.push(providerApi);
			var serviceUrl = apiConstant.api.particularService.url;
			var particularServiceAPI = serviceUrl.replace('{{serviceId}}', serviceId);
			apis.push(particularServiceAPI)

			var promises =[];
			for (var i = 0; i < apis.length; i++) {
				var url = apis[i];
				var method = "GET";
				var headers = {"X-ConsumerId": serviceId};
				promises.push(httpService.apiRequest(url,method,null,headers,null));
			}

			$q.all(promises).then(function(response){
				$scope.instances = response[0].data.instances || [];
				$scope.providers = response[1].data.providers || [];
				$scope.consumers = response[2].data.consumers || [];
				$scope.service = response[3].data.service || [];
			},function(error){
				$scope.instances = [];
				$scope.providers = [];
				$scope.consumers = [];
				$scope.service = [];
			});

			$scope.getInstance = function(){
				$state.go('sc.info.instance')
			};

			$scope.getProvider = function(){
				$state.go('sc.info.provider')
			};

			$scope.getConsumer = function(){
				$state.go('sc.info.consumer')
			};

			$scope.getSchema = function() {
				$state.go('sc.info.schema');
			};

			$scope.getProperties = function() {
				$state.go('sc.info.properties');
			};



			$scope.convertTime = function(timestamp){
				return commonService.timeFormat(timestamp);
			};

			$scope.getActiveTab = function(){
                if($state.current.name == "sc.info.instance"){
                    $scope.selectedTab = 0;
                }
                if($state.current.name == "sc.info.provider"){
                    $scope.selectedTab = 1;
                }
                if($state.current.name == "sc.info.consumer"){
                    $scope.selectedTab = 2;
                }
                if($state.current.name == "sc.info.schema"){
                    $scope.selectedTab = 3;
                }
                if($state.current.name == "sc.info.properties"){
                    $scope.selectedTab = 4;
                }

			}
            $scope.getActiveTab();

}]);
