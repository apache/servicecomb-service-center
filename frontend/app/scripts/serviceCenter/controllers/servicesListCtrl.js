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
	.controller('servicesListController', ['$scope', 'httpService','apiConstant','commonService',
		function($scope, httpService, apiConstant,commonService){

			$scope.appList = 'fetching';
			$scope.serviceList = 'Service List';
			$scope.tableHeaders = [
				{
					'key': 'Name'
				},
				{
					'key': 'Status'
				},
				{
					'key': 'AppId'
				},
				{
					'key': 'Version'
				},
				{
					'key': 'CreatedAt'
				}
			];

			$scope.refreshAppList = function() {
				angular.element(document.querySelector('.fa-refresh')).addClass('fa-spin');
				$scope.getAllServices();
			};

			$scope.getAllServices = function() {
				var url = apiConstant.api.microservice.url;
				var method = apiConstant.api.microservice.method;

				httpService.apiRequest(url, method).then(function(response){
					if(response && response.data && response.data.services){
						$scope.services = [];
						response.data.services.forEach(function(service){
							var servicesList = {
								serviceName: service.serviceName.charAt(0).toUpperCase()+service.serviceName.slice(1).toLowerCase(),
								status: service.status.toLowerCase(),
								appId: service.appId.toLowerCase(),
								version: service.version,
								createdAt: commonService.timeFormat(service.timestamp),
								serviceId: service.serviceId
							};
							$scope.services.push(servicesList);
						});

						if($scope.services.length <= 0){
							$scope.appList = 'empty';
						}
						else {
							$scope.appList = '';
						}
						angular.element(document.querySelector('.fa-refresh')).removeClass('fa-spin');
					}
					else {
						$scope.appList = 'empty';
					}
				},function(error){
					angular.element(document.querySelector('.fa-refresh')).removeClass('fa-spin');
					$scope.appList = 'failed';
				})
			};

			$scope.getAllServices();

	}]);
