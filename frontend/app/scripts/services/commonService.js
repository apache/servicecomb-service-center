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
		.service('commonService', [function(){

			var timeFormat = function(timestamp) {
				var date = new Date(timestamp * 1000);
				var month = date.getMonth()+1;
				var formatedDate = date.getFullYear()+ '-' + month + '-' + date.getDate() + ' ' +date.getHours()+ ':'+ date.getMinutes();
				return formatedDate;
			};

			return {
				timeFormat: timeFormat
			};

}]);
