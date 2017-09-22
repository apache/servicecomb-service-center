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
	.controller('schemaController',['$scope', 'apiConstant', 'httpService', '$stateParams', function($scope, apiConstant, httpService, $stateParams) {
		
		var serviceId = $stateParams.serviceId;
		var schemaApi = apiConstant.api.schema.url;
		var api = schemaApi.replace("{{serviceId}}", serviceId);
		var url = api.replace("{{schemaId}}", serviceId);
		var method = apiConstant.api.schema.method;

		$scope.schema = [];
		// httpService.apiRequest(url, method).then(function(response){
		// 	console.log("got response "+ response);
		// },function(error){
		// 	console.log("error"+ JSON.stringify(error));
		// });

}]);