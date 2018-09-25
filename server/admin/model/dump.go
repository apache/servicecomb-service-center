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
	"github.com/apache/incubator-servicecomb-service-center/version"
)

type Getter interface {
	ForEach(i func(i int, v *KV) bool)
}

type Setter interface {
	SetValue(v *KV)
}

type MicroserviceSlice []*Microservice
type MicroserviceIndexSlice []*MicroserviceIndex
type MicroserviceAliasSlice []*MicroserviceAlias
type TagSlice []*Tag
type MicroServiceRuleSlice []*MicroServiceRule
type MicroServiceDependencyRuleSlice []*MicroServiceDependencyRule
type SummarySlice []*Summary
type InstanceSlice []*Instance

func (s *MicroserviceSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *MicroserviceIndexSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *MicroserviceAliasSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *TagSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *MicroServiceRuleSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *MicroServiceDependencyRuleSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *SummarySlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}
func (s *InstanceSlice) ForEach(f func(i int, v *KV) bool) {
	for i, v := range *s {
		v.KV.Value = v.Value
		if !f(i, v.KV) {
			break
		}
	}
}

func (s *MicroserviceSlice) SetValue(v *KV)      { *s = append(*s, NewMicroservice(v)) }
func (s *MicroserviceIndexSlice) SetValue(v *KV) { *s = append(*s, NewMicroserviceIndex(v)) }
func (s *MicroserviceAliasSlice) SetValue(v *KV) { *s = append(*s, NewMicroserviceAlias(v)) }
func (s *TagSlice) SetValue(v *KV)               { *s = append(*s, NewTag(v)) }
func (s *MicroServiceRuleSlice) SetValue(v *KV)  { *s = append(*s, NewMicroServiceRule(v)) }
func (s *MicroServiceDependencyRuleSlice) SetValue(v *KV) {
	*s = append(*s, NewMicroServiceDependencyRule(v))
}
func (s *SummarySlice) SetValue(v *KV)  { *s = append(*s, NewSummary(v)) }
func (s *InstanceSlice) SetValue(v *KV) { *s = append(*s, NewInstance(v)) }

func NewMicroservice(kv *KV) *Microservice {
	return &Microservice{kv, kv.Value.(*pb.MicroService)}
}
func NewMicroserviceIndex(kv *KV) *MicroserviceIndex {
	return &MicroserviceIndex{kv, kv.Value.(string)}
}
func NewMicroserviceAlias(kv *KV) *MicroserviceAlias {
	return &MicroserviceAlias{kv, kv.Value.(string)}
}
func NewTag(kv *KV) *Tag { return &Tag{kv, kv.Value.(map[string]string)} }
func NewMicroServiceRule(kv *KV) *MicroServiceRule {
	return &MicroServiceRule{kv, kv.Value.(*pb.ServiceRule)}
}
func NewMicroServiceDependencyRule(kv *KV) *MicroServiceDependencyRule {
	return &MicroServiceDependencyRule{kv, kv.Value.(*pb.MicroServiceDependency)}
}
func NewSummary(kv *KV) *Summary { return &Summary{kv, kv.Value.(string)} }
func NewInstance(kv *KV) *Instance {
	return &Instance{kv, kv.Value.(*pb.MicroServiceInstance)}
}

type Cache struct {
	Microservices   MicroserviceSlice               `json:"services,omitempty"`
	Indexes         MicroserviceIndexSlice          `json:"serviceIndexes,omitempty"`
	Aliases         MicroserviceAliasSlice          `json:"serviceAliases,omitempty"`
	Tags            TagSlice                        `json:"serviceTags,omitempty"`
	Rules           MicroServiceRuleSlice           `json:"serviceRules,omitempty"`
	DependencyRules MicroServiceDependencyRuleSlice `json:"dependencyRules,omitempty"`
	Summaries       SummarySlice                    `json:"summaries,omitempty"`
	Instances       InstanceSlice                   `json:"instances,omitempty"`
}

type KV struct {
	Key   string      `json:"key"`
	Rev   int64       `json:"rev"`
	Value interface{} `json:"-"`
}

type Microservice struct {
	*KV
	Value *pb.MicroService `json:"value,omitempty"`
}

type MicroserviceIndex struct {
	*KV
	Value string `json:"value,omitempty"`
}

type MicroserviceAlias struct {
	*KV
	Value string `json:"value,omitempty"`
}

type MicroServiceDependencyRule struct {
	*KV
	Value *pb.MicroServiceDependency `json:"value,omitempty"`
}

type MicroServiceRule struct {
	*KV
	Value *pb.ServiceRule `json:"value,omitempty"`
}

type Summary struct {
	*KV
	Value string `json:"value,omitempty"`
}

type Tag struct {
	*KV
	Value map[string]string `json:"value,omitempty"`
}

type Instance struct {
	*KV
	Value *pb.MicroServiceInstance `json:"value,omitempty"`
}

type DumpRequest struct {
	Options []string
}

type DumpResponse struct {
	Response     *pb.Response        `json:"response,omitempty"`
	Info         *version.VersionSet `json:"info,omitempty"`
	AppConfig    map[string]string   `json:"appConf,omitempty"`
	Environments map[string]string   `json:"environments,omitempty"`
	Cache        *Cache              `json:"cache,omitempty"`
}
