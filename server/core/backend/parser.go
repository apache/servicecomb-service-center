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
package backend

import (
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"strings"
)

// event
type KvEventFunc func(evt KvEvent)

// new
type CreateValueFunc func() interface{}

var (
	newBytes           CreateValueFunc = func() interface{} { return []byte(nil) }
	newString          CreateValueFunc = func() interface{} { return "" }
	newMap             CreateValueFunc = func() interface{} { return make(map[string]string) }
	newService         CreateValueFunc = func() interface{} { return new(pb.MicroService) }
	newInstance        CreateValueFunc = func() interface{} { return new(pb.MicroServiceInstance) }
	newRule            CreateValueFunc = func() interface{} { return new(pb.ServiceRule) }
	newDependencyRule  CreateValueFunc = func() interface{} { return new(pb.MicroServiceDependency) }
	newDependencyQueue CreateValueFunc = func() interface{} { return new(pb.ConsumerDependency) }
)

// parse
type ParseValueFunc func(src []byte, dist interface{}) error

var (
	unParse ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		*d = src
		return nil
	}
	textUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		*d = util.BytesToStringWithNoCopy(src)
		return nil
	}
	mapUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		m := (*d).(map[string]string)
		return json.Unmarshal(src, &m)
	}
	jsonUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		d := dist.(*interface{})
		return json.Unmarshal(src, *d)
	}
)

type Parser interface {
	Unmarshal(src []byte) (interface{}, error)
}

type CommonParser struct {
	NewFunc  CreateValueFunc
	FromFunc ParseValueFunc
}

func (p *CommonParser) Unmarshal(src []byte) (interface{}, error) {
	v := p.NewFunc()
	if err := p.FromFunc(src, &v); err != nil {
		return nil, err
	}
	return v, nil
}

var (
	BytesParser           = &CommonParser{newBytes, unParse}
	StringParser          = &CommonParser{newString, textUnmarshal}
	MapParser             = &CommonParser{newMap, mapUnmarshal}
	ServiceParser         = &CommonParser{newService, jsonUnmarshal}
	InstanceParser        = &CommonParser{newInstance, jsonUnmarshal}
	RuleParser            = &CommonParser{newRule, jsonUnmarshal}
	DependencyRuleParser  = &CommonParser{newDependencyRule, jsonUnmarshal}
	DependencyQueueParser = &CommonParser{newDependencyQueue, jsonUnmarshal}
)

func KvToResponse(kv *KeyValue) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(kv.Key), "/")
}

func GetInfoFromSvcKV(kv *KeyValue) (serviceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromInstKV(kv *KeyValue) (serviceId, instanceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	instanceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromDomainKV(kv *KeyValue) (domain string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}
	domain = keys[l-1]
	return
}

func GetInfoFromProjectKV(kv *KeyValue) (domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}
	domainProject = fmt.Sprintf("%s/%s", keys[l-2], keys[l-1])
	return
}

func GetInfoFromRuleKV(kv *KeyValue) (serviceId, ruleId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	serviceId = keys[l-2]
	ruleId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	return
}

func GetInfoFromTagKV(kv *KeyValue) (serviceId, domainProject string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 3 {
		return
	}
	serviceId = keys[l-1]
	domainProject = fmt.Sprintf("%s/%s", keys[l-3], keys[l-2])
	return
}

func GetInfoFromSvcIndexKV(kv *KeyValue) (key *pb.MicroServiceKey) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 6 {
		return
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func GetInfoFromSchemaSummaryKV(kv *KeyValue) (schemaId string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}

	return keys[l-1]
}

func GetInfoFromSchemaKV(kv *KeyValue) (schemaId string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 2 {
		return
	}

	return keys[l-1]
}

func GetInfoFromDependencyQueueKV(kv *KeyValue) (consumerId, domainProject, uuid string) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 4 {
		return
	}
	consumerId = keys[l-2]
	domainProject = fmt.Sprintf("%s/%s", keys[l-4], keys[l-3])
	uuid = keys[l-1]
	return
}

func GetInfoFromDependencyRuleKV(kv *KeyValue) (key *pb.MicroServiceKey) {
	keys := KvToResponse(kv)
	l := len(keys)
	if l < 5 {
		return
	}
	if keys[l-1] == "*" {
		return &pb.MicroServiceKey{
			Tenant:      fmt.Sprintf("%s/%s", keys[l-5], keys[l-4]),
			Environment: keys[l-2],
			ServiceName: keys[l-1],
		}
	}

	return &pb.MicroServiceKey{
		Tenant:      fmt.Sprintf("%s/%s", keys[l-7], keys[l-6]),
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}
