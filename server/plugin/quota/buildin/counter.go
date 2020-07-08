// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildin

import (
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/quota/counter"
)

var globalCounter = &GlobalCounter{}

func init() {
	counter.RegisterCounter(globalCounter)
}

type GlobalCounter struct {
	ServiceCount  int64
	InstanceCount int64
}

func (c *GlobalCounter) OnCreate(t discovery.Type, domainProject string) {
	switch t {
	case backend.ServiceIndex:
		c.ServiceCount++
	case backend.INSTANCE:
		c.InstanceCount++
	}
}

func (c *GlobalCounter) OnDelete(t discovery.Type, domainProject string) {
	switch t {
	case backend.ServiceIndex:
		if c.ServiceCount == 0 {
			return
		}
		c.ServiceCount--
	case backend.INSTANCE:
		if c.InstanceCount == 0 {
			return
		}
		c.InstanceCount--
	}
}
