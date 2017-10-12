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
	.controller('dashboardController', ['$scope', '$state','apiConstant', 'httpService','$q', '$interval', function($scope, $state, apiConstant, httpService, $q, $interval){

		$scope.labels = ["STRATING", "UP", "DOWN", "OUT OF SERVICE"];
		$scope.colors = ["#00c0ef","#00a65a", "#dd4b39", "#A9A9A9"];
		$scope.servicesChartData = [];
		
		$scope.servicesStatusData = [
			{
				count: 0,
				title: "starting",
				iconName: "fa-spinner"
			},
			{
				count: 0,
				title: "up",
				iconName: "fa-arrow-up"
			},
			{
				count: 0,
				title: "down",
				iconName: "fa-arrow-down"
			},
			{
				count: 0,
				title: "outofservice",
				iconName: "fa-ban"
			}
		];
		$scope.getAllServices = function(){
			$(".loader").show();
			$scope.dashboardInfo = [
				{
					count: 0,
					title: "services",
					iconName: "fa-cog"
				},
				{
					count: 0,
					title: "instances",
					iconName: "fa-cogs"
				},
				{
					count: 0,
					title: "providers",
					iconName: "fa-server"
				},
				{
					count: 0,
					title: "consumers",
					iconName: "fa-users"
				}
			];

			$scope.runningServices = [];
			$scope.stoppedServices = [];
			$scope.startingServices = [];
			$scope.outofserviceServices =[];

			$scope.totalProviders = [];
			$scope.totalConsumers = [];
			
	        var url = apiConstant.api.allServices.url;
	        var method = apiConstant.api.allServices.method;
	        httpService.apiRequest(url, method, null, null).then(function(response){
	            $(".loader").hide();
				if(response && response.data && response.data.allServicesDetail) {
					$scope.dashboardInfo[0].count = response.data.allServicesDetail.length;
					response.data.allServicesDetail.forEach(function(services){
						if(services.microSerivce.status.toLowerCase() == "starting") {
							$scope.startingServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "up"){		
							$scope.runningServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "down"){
							$scope.stoppedServices.push(services);
						}
						if(services.microSerivce.status.toLowerCase() == "outofservice"){
							$scope.outofserviceServices.push(services);
						}

						if(services.instances){
							$scope.dashboardInfo[1].count = $scope.dashboardInfo[1].count + services.instances.length;
						}

						if(services.providers){
							services.providers.forEach(function(provider){
								var isProvider = 0;
								for(var p = 0; p < $scope.totalProviders.length; p++) {
									if(provider.serviceId == $scope.totalProviders[p]) {
										isProvider = 1;
									}
								}
								if(isProvider == 0){
									$scope.totalProviders.push(provider.serviceId);
								}
							})
						}

						if(services.consumers){
							services.consumers.forEach(function(consumer){
								var isConsumer = 0;
								for(var c = 0; c < $scope.totalConsumers.length; c++) {
									if(consumer.serviceId == $scope.totalConsumers[c]) {
										isConsumer = 1;
									}
								}
								if(isConsumer == 0){
									$scope.totalConsumers.push(consumer.serviceId);
								}
							})
						}
					});
					$scope.servicesChartData[0] = $scope.startingServices.length;
					$scope.servicesChartData[1] = $scope.runningServices.length;
					$scope.servicesChartData[2] = $scope.stoppedServices.length;
					$scope.servicesChartData[3] = $scope.outofserviceServices.length;

					$scope.servicesStatusData[0].count = $scope.startingServices.length;
					$scope.servicesStatusData[1].count = $scope.runningServices.length;
					$scope.servicesStatusData[2].count = $scope.stoppedServices.length;
					$scope.servicesStatusData[3].count = $scope.outofserviceServices.length;

					$scope.dashboardInfo[2].count = $scope.totalProviders.length;
					$scope.dashboardInfo[3].count = $scope.totalConsumers.length;
					$scope.timeSince(new Date());

				}
	        },function(error){
	            $(".loader").hide();
	        });
		};
		
		$scope.timeSince = function(date) {
			$scope.updatedTime = "few seconds ago";
		    $interval(function(){
			  var seconds = Math.floor((new Date() - date) / 1000);
			  var interval = Math.floor(seconds / 86400);
			  if (interval > 1) {
			  	$scope.updatedTime = interval + " days ago";
			    return;
			  }
			  interval = Math.floor(seconds / 3600);
			  if (interval > 1) {
			  	$scope.updatedTime = interval + " hours ago";
			    return;
			  }
			  interval = Math.floor(seconds / 60);
			  if (interval > 1) {
			    $scope.updatedTime = interval + " minutes ago";
			    return;
			  }
			  $scope.updatedTime = "1 minute ago";
			},60000)
		};

		$scope.goToServices = function(status){
			$state.go("sc.allServices", {status:status});
		};

		$scope.getAllServices();  
				
}]);