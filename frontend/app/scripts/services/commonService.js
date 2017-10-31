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
		.service('commonService', ['$mdDialog', function($mdDialog){

			var timeFormat = function(timestamp) {
				var date = new Date(timestamp * 1000);
				var month = date.getMonth()+1;
				var formatedDate = date.getFullYear()+ '-' + month + '-' + date.getDate() + ' ' +date.getHours()+ ':'+ date.getMinutes();
				return formatedDate;
			};

			var oneBtnMsg = function(title, message){
				$mdDialog.show({
                    parent: angular.element(document.body),
                    template: `<md-dialog flex="30">
                             <md-toolbar>
                                 <div class="md-toolbar-tools">
                                    <h2>{{ title | translate }}</h2>
                                    <span flex></span>
                                    <md-button class="md-icon-button" ng-click="cancel()">
                                      <md-icon class="glyphicon glyphicon-remove" aria-label="Close dialog"></md-icon>
                                    </md-button>
                                  </div>
                             </md-toolbar>
                             <md-dialog-content>
                                <h3 class="text-center" style="margin-top:15px;">{{ message | translate }}</h3>
                             </md-dialog-content>
                             <md-dialog-actions layout="row">
                                <span flex></span>
                                <md-button ng-click="cancel()">
                                 {{ "close" | translate }}
                                </md-button>
                              </md-dialog-actions>
                            </md-dialog>`,
                    skipHide : true,
                    controller: function($scope, $mdDialog) {
                        $scope.message = message;
                        $scope.title = title;
                        $scope.cancel = function() {
                            $mdDialog.hide();
                        };
                    }
                });
			};

			var twoBtnMsg = function(title, message, callback){
				$mdDialog.show({
                    parent: angular.element(document.body),
                    template: `<md-dialog flex="30">
                             <md-toolbar>
                                 <div class="md-toolbar-tools">
                                    <h2>{{ title | translate }}</h2>
                                    <span flex></span>
                                    <md-button class="md-icon-button" ng-click="no()">
                                      <md-icon class="glyphicon glyphicon-remove" aria-label="Close dialog"></md-icon>
                                    </md-button>
                                  </div>
                             </md-toolbar>
                             <md-dialog-content>
                                <h3 class="text-center" style="margin-top:15px;">{{ message | translate }}</h3>
                             </md-dialog-content>
                             <md-dialog-actions layout="row">
                                <span flex></span>
                                <md-button ng-click="yes()">
                                 {{ "yes" | translate }}
                                </md-button>
                                <md-button ng-click="no()">
                                 {{ "no" | translate }}
                                </md-button>
                              </md-dialog-actions>
                            </md-dialog>`,
                    skipHide : true,
                    controller: function($scope, $mdDialog) {
                        $scope.message = message;
                        $scope.title = title;
                        $scope.yes = function() {
                            $mdDialog.hide();
                            callback("yes");
                        };
                         $scope.no = function() {
                            $mdDialog.hide();
                        };
                    }
                });
			};

			return {
				timeFormat: timeFormat,
				oneBtnMsg: oneBtnMsg,
				twoBtnMsg: twoBtnMsg
			};

}]);
