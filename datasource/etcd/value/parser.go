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

package value

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
	"github.com/go-chassis/cari/discovery"
)

var (
	newService         parser.CreateValueFunc = func() interface{} { return new(discovery.MicroService) }
	newInstance        parser.CreateValueFunc = func() interface{} { return new(discovery.MicroServiceInstance) }
	newRule            parser.CreateValueFunc = func() interface{} { return new(discovery.ServiceRule) }
	newDependencyRule  parser.CreateValueFunc = func() interface{} { return new(discovery.MicroServiceDependency) }
	newDependencyQueue parser.CreateValueFunc = func() interface{} { return new(discovery.ConsumerDependency) }

	ServiceParser         = parser.New(newService, parser.JSONUnmarshal)
	InstanceParser        = parser.New(newInstance, parser.JSONUnmarshal)
	RuleParser            = parser.New(newRule, parser.JSONUnmarshal)
	DependencyRuleParser  = parser.New(newDependencyRule, parser.JSONUnmarshal)
	DependencyQueueParser = parser.New(newDependencyQueue, parser.JSONUnmarshal)
)
