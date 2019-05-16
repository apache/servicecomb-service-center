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
package proto

type SyncMapping []*MappingItem

type MappingItem struct {
	DomainProject string `json:"domain_project"`
	OrgServiceID  string `json:"org_service_id"`
	OrgInstanceID string `json:"org_instance_id"`
	CurServiceID  string `json:"cur_service_id"`
	CurInstanceID string `json:"cur_instance_id"`
}

func (s SyncMapping) OriginIndex(instanceID string) int {
	for index, val := range s {
		if val.OrgInstanceID == instanceID {
			return index
		}
	}
	return -1
}

func (s SyncMapping) CurrentIndex(instanceID string) int {
	for index, val := range s {
		if val.CurInstanceID == instanceID {
			return index
		}
	}
	return -1
}

type ServiceKey struct {
	ServiceName string
	Version     string
}
