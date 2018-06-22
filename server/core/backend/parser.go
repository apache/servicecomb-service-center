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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
)

// event
type KvEventFunc func(evt KvEvent)

// new
type CreateValueFunc func() interface{}

var (
	newBytes          CreateValueFunc = func() interface{} { return []byte(nil) }
	newString         CreateValueFunc = func() interface{} { return "" }
	newMap            CreateValueFunc = func() interface{} { return make(map[string]string) }
	newService        CreateValueFunc = func() interface{} { return new(pb.MicroService) }
	newInstance       CreateValueFunc = func() interface{} { return new(pb.MicroServiceInstance) }
	newRule           CreateValueFunc = func() interface{} { return new(pb.ServiceRule) }
	newDependencyRule CreateValueFunc = func() interface{} { return new(pb.MicroServiceDependency) }
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

type Parser struct {
	c CreateValueFunc
	p ParseValueFunc
}

func (p *Parser) Unmarshal(src []byte) (interface{}, error) {
	v := p.c()
	err := p.p(src, &v)
	return v, err
}

var (
	BytesParser          = &Parser{newBytes, unParse}
	StringParser         = &Parser{newString, textUnmarshal}
	MapParser            = &Parser{newMap, mapUnmarshal}
	ServiceParser        = &Parser{newService, jsonUnmarshal}
	InstanceParser       = &Parser{newInstance, jsonUnmarshal}
	RuleParser           = &Parser{newRule, jsonUnmarshal}
	DependencyRuleParser = &Parser{newDependencyRule, jsonUnmarshal}
)
