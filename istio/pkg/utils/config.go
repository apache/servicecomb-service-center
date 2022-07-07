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

package utils

import "time"

// Params under this const is used for service center controller
const (
	// The name used for all created service center Watcher services.
	WATCHER_SVC_NAME string = "servicecenter-watcher"
	// Reserved service name used by ServiceCenter, will be ignored.
	// This is the default name of the ServiceCenter in an etcd-based registry
	// (https://github.com/apache/servicecomb-service-center/blob/6f26aaa7698691d40e17c6644ac71d51b6770772/examples/etcd_data_struct.yaml#L7).
	SERVICECENTER_ETCD_NAME string = "SERVICECENTER"
	// Reserved service name used by ServiceCenter, will be ignored.
	// This is the default name of the ServiceCenter in a MongoDB-based registry
	// (https://github.com/apache/servicecomb-service-center/blob/6f26aaa7698691d40e17c6644ac71d51b6770772/examples/mongodb_data_struct.yaml#L27).
	SERVICECENTER_MONGO_NAME string = "SERVICE-CENTER"
	// Time in seconds to wait before syncing new services from service center registry.
	PULL_INTERVAL time.Duration = time.Second * 5
)

// Params under this const is used for istio controller
const (
	// The default k8s namespace to create new Istio ServiceEntry(s) synced from service center in.
	ISTIO_SYSTEM = "istio-system"
	// PUSH_DEBOUNCE_INTERVAL is the delay added to events to wait after a registry event for debouncing.
	// This will delay the push by at least this interval, plus the time getting subsequent events.
	// If no change is detected the push will happen, otherwise we'll keep delaying until things settle.
	PUSH_DEBOUNCE_INTERVAL time.Duration = 1 * time.Second
	// PUSH_DEBOUNCE_MAX_INTERVAL is the maximum time to wait for events while debouncing.
	// Defaults to 10 seconds. If events keep showing up with no break for this time, we'll trigger a push.
	PUSH_DEBOUNCE_MAX_INTERVAL time.Duration = 10 * time.Second
	// PUSH_DEBOUNCE_MAX_EVENTS is maximum number events in the debouncing waiting queue, once it is rached
	// the push function will be fired
	PUSH_DEBOUNCE_MAX_EVENTS int = 1000
)
