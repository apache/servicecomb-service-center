/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
'use strict';
angular.module('serviceCenter')
    .directive('tableData', ['$rootScope', '$mdDialog', '$state', '$translate', function($rootScope, $mdDialog, $state, $translate) {
        return {
            restrict: 'E',
            scope: {
                heading: '=heading',
                headers: '=headers',
                data: '=data',
                pagination: '=pagination',
                refresh: '=refresh',
                removes: '=removes',
                dialogClose: '=dialogClose',
                buttons: '=buttons',
                enableSearch: '=search',
                appList: '=appList',
                apiInfo: '=apiInfo',
                searchFn: '=searchFn',
                statusList: '=statusList'
            },
            templateUrl: 'scripts/views/tableData.html',
            link: function(scope) {
                /*
                	If the table has buttons with functions declared in the controller then the functions must be defined first
                	and then the buttons configured. Seems like some scope issue need to figure out a better solution
                */
                scope.paginationQuery = {
                    order: 'name',
                    limit: 10,
                    page: 1
                };

                $rootScope.$on('$translateChangeSuccess', function() {
                    scope.changePaginationLabel();
                });

                scope.changePaginationLabel = function() {
                    scope.paginationLabel = {
                        page: $translate.instant('page'),
                        rowsPerPage: $translate.instant('rowsPerPage'),
                        of: $translate.instant('of')
                    }
                }

                scope.changePaginationLabel()

                scope.paginationOptions = {
                    rowSelection: false,
                    multiSelect: false,
                    autoSelect: false,
                    decapitate: false,
                    largeEditDialog: false,
                    boundaryLinks: true,
                    limitSelect: true,
                    pageSelect: true
                };

                scope.filter = {
                    options: {
                        debounce: 500
                    },
                    status: scope.statusList ? scope.statusList[0].id : []
                };

                if (scope.statusList && scope.statusList.length > 0) {
                    scope.filter.status = $state.params.status ? $state.params.status : scope.statusList[0].id;
                }

                scope.showSearch = false;
                scope.searchClose = function() {
                    scope.showSearch = false;
                    scope.filter.search = "";
                    scope.fnSearch(scope.filter.search, scope.filter.status);
                };

                scope.searchOpen = function() {
                    scope.showSearch = true;
                };

                scope.reload = function() {
                    scope.showSearch = false;
                    scope.filter.search = "";
                    scope.refresh();
                    if (scope.statusList && scope.statusList.length > 0) {
                        scope.filter.status = $state.params.status ? $state.params.status : scope.statusList[0].id;
                    }
                };

                scope.close = function() {
                    $mdDialog.cancel();
                };

                scope.remove = function(serviceId, instances) {
                    scope.removes(serviceId, instances)
                };

                scope.fnSearch = function(searchData, status) {
                    scope.filter.status = status !== undefined ? status : scope.statusList[0].id;
                    scope.searchFn(searchData, status);
                }

            }
        };
    }]);
