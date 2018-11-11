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

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
)

type Cache []discovery.Cache

func (c Cache) Name() string {
	if len(c) == 0 {
		return ""
	}
	return c[0].Name()
}

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
	exists := make(map[string]struct{})
	for _, cache := range c {
		var tmp []*discovery.KeyValue
		if l := cache.GetAll(&tmp); l == 0 {
			continue
		}
		s += c.append(tmp, arr, exists)
	}
	return
}
func (c Cache) GetPrefix(prefix string, arr *[]*discovery.KeyValue) (s int) {
	exists := make(map[string]struct{})
	for _, cache := range c {
		var tmp []*discovery.KeyValue
		if l := cache.GetPrefix(prefix, &tmp); l == 0 {
			continue
		}
		s += c.append(tmp, arr, exists)
	}
	return
}

func (c Cache) append(tmp []*discovery.KeyValue, arr *[]*discovery.KeyValue,
	exists map[string]struct{}) (s int) {
	for _, kv := range tmp {
		key := util.BytesToStringWithNoCopy(kv.Key)
		if _, ok := exists[key]; ok {
			continue
		}
		exists[key] = struct{}{}
		if arr != nil {
			*arr = append(*arr, kv)
		}
		s += 1
	}
	return
}

func (c Cache) ForEach(iter func(k string, v *discovery.KeyValue) (next bool)) {
	exists := make(map[string]struct{})
	for _, cache := range c {
		cache.ForEach(func(k string, v *discovery.KeyValue) bool {
			if _, ok := exists[k]; ok {
				return true
			}
			exists[k] = struct{}{}
			return iter(k, v)
		})
	}
}
func (c Cache) Put(k string, v *discovery.KeyValue) { return }
func (c Cache) Remove(k string)                     { return }
