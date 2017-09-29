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
angular.module('serviceCenter.dashboard', [])
	.controller('dashboardController', ['$scope', '$state','apiConstant', 'httpService','$q', function($scope, $state, apiConstant, httpService,$q){

		$scope.servicesData = [];
		$scope.runningServices = [];
		$scope.stoppedServices = [];
		$scope.startingServices = [];
		$scope.outofserviceServices =[];

		$scope.totalServices = 0;
		$scope.totalInstances = 0;	
		$scope.totalProviders = 0;
		$scope.totalConsumers = 0;

		$scope.labels = ["UP", "DOWN", "STRATING", "OUT OF SERVICE"];
		$scope.colors = ["#5ecc49","#d50000", "#e6e600", "#bcc2c9"];

		$scope.getAllServices = function(){
			$(".loader").show();
	        var url = apiConstant.api.allServices.url;
	        var method = apiConstant.api.allServices.method;
	        httpService.apiRequest(url, method, null, null).then(function(response){
	            $(".loader").hide();
				if(response && response.data && response.data.allServicesDetail) {
					$scope.totalServices = response.data.allServicesDetail.length;
					response.data.allServicesDetail.forEach(function(services){
						if(services.microSerivce.status.toLowerCase() == "starting") {
							$scope.startingServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "up"){		
							console.log("hi")
							$scope.runningServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "down"){
							$scope.stoppedServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "outofservice"){
							$scope.outofserviceServices.push(services);
						}

						if(services.instances){
							$scope.totalInstances = $scope.totalInstances + services.instances.length;
						}

						if(services.providers){
							$scope.totalProviders = $scope.totalInstances + services.providers.length;
						}

						if(services.consumers){
							$scope.totalConsumers = $scope.totalInstances + services.consumers.length;
						}
					});
					$scope.servicesData[0] = $scope.runningServices.length;
					$scope.servicesData[1] = $scope.stoppedServices.length;
					$scope.servicesData[2] = $scope.startingServices.length;
					$scope.servicesData[3] = $scope.outofserviceServices.length;
				}
	        },function(error){
	            $(".loader").hide();
	        });
		};
	
		$scope.getAllServices();  
				
}]);