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
angular.module('serviceCenter.sc')
	.controller('schemaController',['$scope', 'apiConstant', 'httpService', '$stateParams', 'servicesList', '$q', '$mdDialog', 'YAML',
		function($scope, apiConstant, httpService, $stateParams, servicesList, $q, $mdDialog, YAML) {
		
		var serviceId = $stateParams.serviceId;
		$scope.schemaName = [];
		if(servicesList && servicesList.data && servicesList.data.services){
			servicesList.data.services.forEach(function(services){
	            if(services.serviceId == serviceId){
	                $scope.schemaName = services.schemas;
	            }
        	});
		}
		var addresses = [];
		var instances = [];
		$scope.instanceDetails = function(){
			var instanceUrl = apiConstant.api.instances.url;
			var instanceApi = instanceUrl.replace('{{serviceId}}', serviceId);
			var instanceMethod = apiConstant.api.instances.method;
			var instanceHeaders = {"X-ConsumerId": serviceId};
			httpService.apiRequest(instanceApi, instanceMethod, null, instanceHeaders, "nopopup").then(function(response){
			  if(response && response.data && response.data.instances) {
					for(var i = 0; i < response.data.instances.length; i++){
						instances.push(response.data.instances[i].hostName);
						for(var j = 0; j< response.data.instances[i].endpoints.length; j++){
							addresses[i] = [];
							addresses[i].push(response.data.instances[i].endpoints[j])
						}
					}
			  }
			  else {
			  	addresses = [];
			  }
			},function(error){
				addresses = [];
			});
		}
		$scope.instanceDetails();

		$scope.schema = [];
		$scope.testSchema = function(selectedSchema) {
			$mdDialog.show({
		      controller: function ($scope, $mdDialog, apiConstant, httpService) {
				    $scope.showSchema = false;

				    $scope.instances = instances;
				    $scope.selectedInstance =  instances[0] || '';

				    $scope.addresses = addresses[0];
				    $scope.selectedAddress = addresses[0][0] || '';

				    $scope.setInstance = function(instance) {
				    	for(var i = 0; i < $scope.instances.length; i++){
				    		if(instance == $scope.instances[i]){
				    			$scope.selectedInstance = instance[i];
				    			$scope.addresses = addresses[i];
				    			$scope.selectedAddress = addresses[i][0];
				    		}
				    	}

				    };

				    $scope.setAddress = function(address) {
				    	$scope.selectedAddress = address;
				    };

				    $scope.getSchema = function(){
				    	var schemaApi = apiConstant.api.schema.url;
						var api = schemaApi.replace("{{serviceId}}", serviceId);
						var url = api.replace("{{schemaId}}", selectedSchema);
						var method = apiConstant.api.schema.method;
						var headers = {"X-ConsumerId": serviceId};
						httpService.apiRequest(url, method, null, headers, "nopopup").then(function(response){
							$(".loader").hide();
							if(response && response.data && response.data.schema){
								if($scope.selectedAddress.indexOf("rest") != -1){
				    				var ip = $scope.selectedAddress.substring(7, $scope.selectedAddress.length);
				    			}
				    			if($scope.selectedAddress.indexOf("highway") != -1){
				    				var ip = $scope.selectedAddress.substring(10, $scope.selectedAddress.length);
				    			}
								var schema = response.data.schema;
								var json = YAML.parse(schema);
								json.basePath = "/testSchema";
								json.instanceIP = ip;
								json.schemaName = selectedSchema;
								var yamlString = YAML.stringify(json);
								$scope.testSchema = yamlString;
								$scope.showSchema = true;
							}
						},function(error) {
							$(".loader").hide();
							$scope.showSchema = true;
						});
				    }

					$scope.hide = function() {
				      $mdDialog.hide();
				    };

				    $scope.cancel = function() {
				      $mdDialog.cancel();
				    };
			  },
		      templateUrl: 'scripts/modules/serviceCenter/views/testSchema.html',
		      parent: angular.element(document.body),
		      clickOutsideToClose:true,
		      fullscreen: false
		    });
		};
}]);