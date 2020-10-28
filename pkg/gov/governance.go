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

package gov

//GovernancePolicy is a unified struct
//all governance policy must extend this struct
//Name is the policy name, for example: "rate-limit-payment-api"
//MD is metadata.
type GovernancePolicy struct {
	Name string            `json:"name,omitempty"`
	MD   map[string]string `json:"metadata,omitempty"`
}

//LoadBalancer define policy and fault tolerant policy
type LoadBalancer struct {
	*GovernancePolicy
	Spec *LBSpec `json:"spec,omitempty"`
}
type LBSpec struct {
	MarkerName string         `json:"match"`
	RetrySame  int            `json:"retrySame,omitempty"`
	RetryNext  int            `json:"retryNext,omitempty"`
	Bo         *BackOffPolicy `json:"backoff,omitempty"`
}
type BackOffPolicy struct {
	InitialInterval int `json:"initInterval"`
	MaxInterval     int `json:"maxInterval"`
}
