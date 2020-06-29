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

package counter

import (
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
)

var counters = Counters{}

type Counter interface {
	OnCreate(t discovery.Type, domainProject string)
	OnDelete(t discovery.Type, domainProject string)
}

type Counters []Counter

func (cs Counters) OnCreate(t discovery.Type, domainProject string) {
	for _, c := range cs {
		c.OnCreate(t, domainProject)
	}
}

func (cs Counters) OnDelete(t discovery.Type, domainProject string) {
	for _, c := range cs {
		c.OnDelete(t, domainProject)
	}
}

func RegisterCounter(c Counter) {
	counters = append(counters, c)
}

func GetCounters() Counters {
	return counters
}
