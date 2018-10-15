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

package aggregate

import "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/discovery"

type Cache []discovery.Cache

func (c Cache) Name() string { return c[0].Name() }
func (c Cache) Size() (s int) {
	for _, cache := range c {
		s += cache.Size()
	}
	return s
}
func (c Cache) Get(k string) (kv *discovery.KeyValue) {
	for _, cache := range c {
		if kv = cache.Get(k); kv != nil {
			return
		}
	}
	return
}
func (c Cache) GetAll(arr *[]*discovery.KeyValue) (s int) {
	for _, cache := range c {
		s += cache.GetAll(arr)
	}
	return
}
func (c Cache) GetPrefix(prefix string, arr *[]*discovery.KeyValue) (s int) {
	for _, cache := range c {
		s += cache.GetPrefix(prefix, arr)
	}
	return
}
func (c Cache) ForEach(iter func(k string, v *discovery.KeyValue) (next bool)) {
	for _, cache := range c {
		cache.ForEach(iter)
	}
}
func (c Cache) Put(k string, v *discovery.KeyValue) { return }
func (c Cache) Remove(k string)                     { return }
