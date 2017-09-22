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
	.controller('dashboardController', ['$scope','servicesList', 'instancesList', '$state', function($scope, servicesList, instancesList, $state){

		$scope.runningServices = [];
		$scope.stoppedServices = [];

		$scope.runningInstances = [];
		$scope.stoppedInstances = [];

		$scope.totalServices = servicesList.length;
		$scope.totalInstances = instancesList.length;
		
		servicesList.forEach(function(service){
			if(service.status.toLowerCase() === "up"){
				$scope.runningServices.push(service);
			}
			if(service.status.toLowerCase() === "down"){
				$scope.stoppedServices.push(service);
			}
		});

		instancesList.forEach(function(instance){
			if(instance.status.toLowerCase() === "up"){
				$scope.runningInstances.push(instance);
			}
			if(instance.status.toLowerCase() === "down"){
				$scope.stoppedInstances.push(instance);
			}
		});

		$scope.servicesData = [];
		$scope.servicesData[0] = $scope.runningServices.length;
		$scope.servicesData[1] = $scope.stoppedServices.length;
		$scope.labels = ["UP", "DOWN"];

		$scope.instancesData = [];
		$scope.instancesData[0] = $scope.runningInstances.length;
		$scope.instancesData[1] = $scope.stoppedInstances.length;
		
		$scope.getServices = function(){
			$state.go('sc.allServices');
		};

		$scope.getInstances = function(){
			$state.go('sc.allInstances');
		};

		
		 
}]);