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

package dao

import "context"

//TODO save to etcd
//TODO now simply write dead code "*" to map all other API except account and role to service, should define resource for every API in future
var resourceMap = map[string]string{
	"/v4/account": "account",
	"/v4/role":    "account",
	"*":           "service",
}

func GetResource(ctx context.Context, API string) string {
	r, ok := resourceMap[API]
	if !ok {
		return resourceMap["*"]
	}
	return r
}
