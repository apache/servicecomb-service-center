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
	.controller('serviceInfoController', ['$scope', 'httpService', 'commonService', '$q', 'apiConstant', '$state', '$stateParams','currentService',
		function($scope, httpService, commonService, $q, apiConstant, $state, $stateParams, currentService){

			var serviceId = $stateParams.serviceId;
			$scope.currentServiceDetail = currentService;

			var apis = [];
			var instanceUrl = apiConstant.api.instances.url;
			var instanceApi = instanceUrl.replace('{{serviceId}}', serviceId);
			apis.push(instanceApi);
			var consumerUrl = apiConstant.api.consumer.url;
			var consumerApi = consumerUrl.replace('{{consumerId}}', serviceId);
			apis.push(consumerApi);
			var providerUrl = apiConstant.api.provider.url;
			var providerApi = providerUrl.replace('{{proviserId}}', serviceId);
			apis.push(providerApi);

			var promises =[];
			for (var i = 0; i < apis.length; i++) {
				var url = apis[i];
				var method = "GET";
				var headers = {"X-ConsumerId": serviceId};
				promises.push(httpService.apiRequest(url,method,null,headers));
			}

			$q.all(promises).then(function(response){
				$scope.instances = response[0].data.instances;
				$scope.providers = response[1].data.providers;
				$scope.consumers = response[2].data.consumers;
			},function(error){
				$scope.instances = [];
				$scope.providers = [];
				$scope.consumers = [];
			});

			$scope.showInstance = function(){
				$state.go('sc.info.instance')
			};

			$scope.showProvider = function(){
				$state.go('sc.info.provider')
			};

			$scope.showConsumer = function(){
				$state.go('sc.info.consumer')
			};

			$scope.convertTime = function(timestamp){
				return commonService.timeFormat(timestamp);
			};

}]);
