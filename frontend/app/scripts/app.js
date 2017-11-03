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

angular.module('serviceCenter', ['ngAnimate', 'ngMaterial', 'ngAria', 'ngMessages', 'ngResource', 'ngSanitize', 'ui.router', 'pascalprecht.translate', 'serviceCenter.router', 
  'serviceCenter.dashboard', 'serviceCenter.sc', 'md.data.table', 'chart.js', 'swaggerUi', 'mmumshad.yamljs'])
  .config(['$translateProvider', 'english', 'chinese', function($translateProvider, english, chinese) {
        $translateProvider.useSanitizeValueStrategy(null);
        
        $translateProvider.translations('en', english);
        $translateProvider.translations('cz', chinese);
  
        var lang = "";
        if(localStorage.getItem("lang") && localStorage.getItem("lang")!= ''){
            lang= localStorage.getItem("lang");
        }
        else if (navigator.language) {
            lang = navigator.language.indexOf("zh") > -1 ? "cz" : "en";
        } else {
            lang = navigator.userLanguage.indexOf("zh") > -1 ? "cz" : "en";
        }

        $translateProvider.preferredLanguage(lang);
    }])
  .config(['$httpProvider','$injector', '$compileProvider', function($httpProvider,$injector, $compileProvider) {
        $httpProvider.defaults.useXDomain = true;
        delete $httpProvider.defaults.headers.common['X-Requested-With'];

        $injector.invoke(['$qProvider', function($qProvider) {
            $qProvider.errorOnUnhandledRejections(false);
        }]);

        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|ftp|mailto|tel|file|blob):/)

    }])
  .config(function($mdThemingProvider) { 
    $mdThemingProvider.theme('default')
      .primaryPalette('indigo', {
        'default': '400',
        'hue-1': '100',
        'hue-2': '600', 
        'hue-3': 'A100' 
      })
      .accentPalette('purple', {
        'default': '200'
      });
  });
