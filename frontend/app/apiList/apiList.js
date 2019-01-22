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
	.constant('apiConstant', {
            // v4 api's
		api : {
            instances: {
                url: 'v4/default/registry/microservices/{{serviceId}}/instances',
                method: 'GET'
            },
            consumer: {
                url: 'v4/default/registry/microservices/{{consumerId}}/providers',
                method: 'GET'
            },
            particularService: {
                url: 'v4/default/registry/microservices/{{serviceId}}',
                method: 'GET'
            },
            provider: {
                url: 'v4/default/registry/microservices/{{providerId}}/consumers',
                method: 'GET'
            },
            schema: {
                url: 'v4/default/registry/microservices/{{serviceId}}/schemas/{{schemaId}}',
                method: 'GET'
            },
            allServices: {
                url: 'v4/default/govern/microservices?options=instances,dependencies&withShared=true',
                method: 'GET'
            },
            deleteService: {
                url: 'v4/default/registry/microservices/{{serviceId}}?force=true',
                method: 'DELETE'
            }
        }
//       v3 api's
//		api : {
//			microservice: {
//				url: 'registry/v3/microservices',
//				method: 'GET'
//			},
//			instances: {
//				url: 'registry/v3/microservices/{{serviceId}}/instances?tags=',
//				method: 'GET'
//			},
//			consumer: {
//				url: 'registry/v3/microservices/{{consumerId}}/providers',
//				method: 'GET'
//			},
//			provider: {
//				url: 'registry/v3/microservices/{{providerId}}/consumers',
//				method: 'GET'
//			},
//			schema: {
//				url: 'registry/v3/microservices/{{serviceId}}/schemas/{{schemaId}}',
//				method: 'GET'
//			},
//			allServices: {
//				url: 'registry/v3/govern/services?options=all',
//				method: 'GET'
//			},
//          deleteService: {
//              url: 'registry/v3/microservices/{{serviceId}}',
//              method: 'DELETE'
//          }
//		}
	});
