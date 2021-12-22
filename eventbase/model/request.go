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

package model

// GetTombstoneRequest contains tombstone get request params
type GetTombstoneRequest struct {
	Project      string `json:"project,omitempty" yaml:"project,omitempty"`
	Domain       string `json:"domain,omitempty" yaml:"domain,omitempty"`
	ResourceType string `json:"resource_type,omitempty" yaml:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty" yaml:"resource_id,omitempty"`
}

// ListTaskRequest contains task list request params
type ListTaskRequest struct {
	Domain       string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Project      string `json:"project,omitempty" yaml:"project,omitempty"`
	Action       string `json:"action,omitempty" yaml:"action,omitempty"`
	Status       string `json:"status,omitempty" yaml:"status,omitempty"`
	ResourceType string `json:"resource_type,omitempty" yaml:"resource_type,omitempty"`
}

// ListTombstoneRequest contains tombstone list request params
type ListTombstoneRequest struct {
	Domain          string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Project         string `json:"project,omitempty" yaml:"project,omitempty"`
	ResourceType    string `json:"resource_type,omitempty" yaml:"resource_type,omitempty"`
	BeforeTimestamp int64  `json:"before_timestamp,omitempty" yaml:"before_timestamp,omitempty"`
}
