'use strict';
angular.module('serviceCenter')
	.controller('baseController', ['$scope', '$state', '$stateParams',
		function($scope, $state, $stateParams){

			$scope.$state = $state;
			$scope.$stateParams = $stateParams;
			
}]);	