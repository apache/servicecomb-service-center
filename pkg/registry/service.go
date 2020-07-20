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

package registry

type MicroService struct {
	ServiceId    string             `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	AppId        string             `protobuf:"bytes,2,opt,name=appId" json:"appId,omitempty"`
	ServiceName  string             `protobuf:"bytes,3,opt,name=serviceName" json:"serviceName,omitempty"`
	Version      string             `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
	Description  string             `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	Level        string             `protobuf:"bytes,6,opt,name=level" json:"level,omitempty"`
	Schemas      []string           `protobuf:"bytes,7,rep,name=schemas" json:"schemas,omitempty"`
	Paths        []*ServicePath     `protobuf:"bytes,10,rep,name=paths" json:"paths,omitempty"`
	Status       string             `protobuf:"bytes,8,opt,name=status" json:"status,omitempty"`
	Properties   map[string]string  `protobuf:"bytes,9,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Timestamp    string             `protobuf:"bytes,11,opt,name=timestamp" json:"timestamp,omitempty"`
	Providers    []*MicroServiceKey `protobuf:"bytes,12,rep,name=providers" json:"providers,omitempty"`
	Alias        string             `protobuf:"bytes,13,opt,name=alias" json:"alias,omitempty"`
	LBStrategy   map[string]string  `protobuf:"bytes,14,rep,name=LBStrategy" json:"LBStrategy,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ModTimestamp string             `protobuf:"bytes,15,opt,name=modTimestamp" json:"modTimestamp,omitempty"`
	Environment  string             `protobuf:"bytes,16,opt,name=environment" json:"environment,omitempty"`
	RegisterBy   string             `protobuf:"bytes,17,opt,name=registerBy" json:"registerBy,omitempty"`
	Framework    *FrameWorkProperty `protobuf:"bytes,18,opt,name=framework" json:"framework,omitempty"`
}

// 删除服务请求
type DelServicesRequest struct {
	ServiceIds []string `protobuf:"bytes,1,rep,name=serviceIds" json:"serviceIds,omitempty"`
	Force      bool     `protobuf:"varint,2,opt,name=force" json:"force,omitempty"`
}
