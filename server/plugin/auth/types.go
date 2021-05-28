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

package auth

// ResourceScope is the resource scope parsed from request
type ResourceScope struct {
	Type string
	// Labels is a map used to filter resource permissions during pre verification.
	// If a key of permission set is missing in the Labels, pre verification will pass this key
	Labels []map[string]string
	// Verb is the apply resource action, e.g. "get", "create"
	Verb string
}
