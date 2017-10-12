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
	.controller('schemaController',['$scope', 'apiConstant', 'httpService', '$stateParams', 'servicesList', '$q', '$mdDialog',
		function($scope, apiConstant, httpService, $stateParams, servicesList, $q, $mdDialog) {
		
		var serviceId = $stateParams.serviceId;
		$scope.schemaName = [];
		if(servicesList && servicesList.data && servicesList.data.services){
			servicesList.data.services.forEach(function(services){
	            if(services.serviceId == serviceId){
	                $scope.schemaName = services.schemas;
	            }
        	});
		}
		
		$scope.schema = [];

		$scope.testSchema = function(selectedSchema) {
			$mdDialog.show({
		      controller: function ($scope, $mdDialog, apiConstant, httpService) {
				    $scope.hide = function() {
				      $mdDialog.hide();
				    };

				    $scope.cancel = function() {
				      $mdDialog.cancel();
				    };

				    var schemaApi = apiConstant.api.schema.url;
					var api = schemaApi.replace("{{serviceId}}", serviceId);
					var url = api.replace("{{schemaId}}", selectedSchema);
					var method = apiConstant.api.schema.method;
					var headers = {"X-ConsumerId": serviceId};
					httpService.apiRequest(url, method, null, headers, "nopopup").then(function(response){
						$(".loader").hide();
						if(response && response.data){
							$scope.testSchema = response.data.schema;
						}else {
						}
					},function(error) {
						 	$(".loader").hide();
					});
			  },
		      templateUrl: 'scripts/modules/serviceCenter/views/testSchema.html',
		      parent: angular.element(document.body),
		      clickOutsideToClose:true,
		      fullscreen: false
		    });
		};
		
}]);