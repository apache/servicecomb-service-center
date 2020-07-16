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

package rbacframe

//as a user, he only understand resource of this system,
//but to decouple authorization code from business code,
//a middleware should handle all the authorization logic, and this middleware only understand rest API,
//a resource mapping helps to do it.
var resourceMap = map[string]string{}

func GetResource(api string) string {
	r, ok := resourceMap[api]
	if !ok {
		return resourceMap["*"]
	}
	return r
}

//MapResource save the mapping from api to resource
func MapResource(api, resource string) {
	resourceMap[api] = resource
}
