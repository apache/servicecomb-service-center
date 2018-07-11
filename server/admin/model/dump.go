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

import (
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
)

type Cache struct {
	Microservices   []Microservice               `json:"services,omitempty"`
	Indexes         []MicroserviceIndex          `json:"serviceIndexes,omitempty"`
	Aliases         []MicroserviceAlias          `json:"serviceAliases,omitempty"`
	Tags            []Tag                        `json:"serviceTags,omitempty"`
	Rules           []MicroServiceRule           `json:"serviceRules,omitempty"`
	DependencyRules []MicroServiceDependencyRule `json:"dependencyRules,omitempty"`
	Summaries       []Summary                    `json:"summaries,omitempty"`
	Instances       []Instance                   `json:"instances,omitempty"`
}

type KV struct {
	Key string `json:"key"`
	Rev int64  `json:"rev"`
}

type Microservice struct {
	KV
	Value *pb.MicroService `json:"value,omitempty"`
}

type MicroserviceIndex struct {
	KV
	Value string `json:"value,omitempty"`
}

type MicroserviceAlias struct {
	KV
	Value string `json:"value,omitempty"`
}

type MicroServiceDependencyRule struct {
	KV
	Value *pb.MicroServiceDependency `json:"value,omitempty"`
}

type MicroServiceRule struct {
	KV
	Value *pb.ServiceRule `json:"value,omitempty"`
}

type Summary struct {
	KV
	Value string `json:"value,omitempty"`
}

type Tag struct {
	KV
	Value map[string]string `json:"value,omitempty"`
}

type Instance struct {
	KV
	Value *pb.MicroServiceInstance `json:"value,omitempty"`
}

type DumpRequest struct {
	Options []string
}

type DumpResponse struct {
	Response *pb.Response `json:"response,omitempty"`
	Cache    Cache        `json:"cache,omitempty"`
}
