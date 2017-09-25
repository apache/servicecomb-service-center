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
	.controller('dashboardController', ['$scope', '$state','apiConstant', 'httpService','$q', function($scope, $state, apiConstant, httpService,$q){

		$scope.servicesData = [];
		$scope.runningServices = [];
		$scope.stoppedServices = [];

		$scope.instancesData = [];
		$scope.runningInstances = [];
		$scope.stoppedInstances = [];	

		$scope.totalServices = 0;
		$scope.totalInstances = 0;	

		$scope.labels = ["UP", "DOWN"];
		var promises = [];
		
		$scope.getServices = function(){
			$(".loader").show();
	        var url = apiConstant.api.microservice.url;
	        var method = apiConstant.api.microservice.method;
	        httpService.apiRequest(url,method,null,null).then(function(response){
	            $(".loader").hide();
	            if(response && response.data && response.data.services){
					$scope.totalServices = response.data.services.length;
	            	response.data.services.forEach(function(service){
						if(service.status.toLowerCase() === "up"){
							$scope.runningServices.push(service);
						}
						if(service.status.toLowerCase() === "down"){
							$scope.stoppedServices.push(service);
						}
		   			});
					$scope.servicesData[0] = $scope.runningServices.length;
					$scope.servicesData[1] = $scope.stoppedServices.length;
	            }
	        },function(error){
	            $(".loader").hide();
	        });
		};

		$scope.getServices();

		$scope.getInstances = function(){
			$(".loader").show();
			var url = apiConstant.api.microservice.url;
            var method = apiConstant.api.microservice.method;
            httpService.apiRequest(url,method,null, null, "nopopup").then(function(response){
	            $(".loader").hide();
	            if(response && response.data && response.data.services){
	                for (var i = 0; i < response.data.services.length; i++) {
	                    var api = apiConstant.api.instances.url;
	                    var url = api.replace("{{serviceId}}", response.data.services[i].serviceId);
	                    var method = apiConstant.api.instances.method;
	                    var headers = {"X-ConsumerId": response.data.services[i].serviceId};

	                    promises.push(httpService.apiRequest(url,method,null,headers,"nopopup"));
	                }

	                $q.all(promises).then(function(response){
			            if(response && response[0].data && response[0].data.instances){
							$scope.totalInstances = response[0].data.instances.length;
							response[0].data.instances.forEach(function(instance){
								if(instance.status.toLowerCase() === "up"){
									$scope.runningInstances.push(instance);
								}
								if(instance.status.toLowerCase() === "down"){
									$scope.stoppedInstances.push(instance);
								}
							});
							$scope.instancesData[0] = $scope.runningInstances.length;
							$scope.instancesData[1] = $scope.stoppedInstances.length;
			            }
			        },function(error){
			        	$(".loader").hide();
			        });

	            }
	        },function(error){
	          	 $(".loader").hide();
	        });
	            
		};
		
		$scope.getInstances();
		
		$scope.getServices = function(){
			$state.go('sc.allServices');
		};

		$scope.getInstances = function(){
			$state.go('sc.allInstances');
		};

}]);