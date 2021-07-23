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
	Name       string    `json:"name,omitempty"`
	ID         string    `json:"id,omitempty"`
	Status     string    `json:"status,omitempty"`
	CreatTime  int64     `json:"creatTime,omitempty"`
	UpdateTime int64     `json:"updateTime,omitempty"`
	Selector   *Selector `json:"selector,omitempty"`
}

//DisplayData define display data
type DisplayData struct {
	Policies   []*Policy `json:"policies,omitempty"`
	MatchGroup *Policy   `json:"matchGroup,omitempty"`
}

//Policy define policy and fault tolerant policy
type Policy struct {
	*GovernancePolicy
	Kind string                 `json:"kind,omitempty"`
	Spec map[string]interface{} `json:"spec,omitempty"`
}

type Selector struct {
	App         string `json:"app,omitempty"`
	Environment string `json:"environment,omitempty"`
}

type LBSpec struct {
	MarkerName string         `json:"match"`
	RetrySame  int            `json:"retrySame,omitempty"`
	RetryNext  int            `json:"retryNext,omitempty"`
	Bo         *BackOffPolicy `json:"backoff,omitempty"`
	Alias      string         `json:"alias"`
}
type BackOffPolicy struct {
	InitialInterval int `json:"initInterval"`
	MaxInterval     int `json:"maxInterval"`
}
