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
	"context"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"time"
)

type ConflictChecker struct {
	Cache              discovery.CacheReader
	ConflictHandleFunc func(origin, conflict *discovery.KeyValue)
}

func (c *ConflictChecker) Run(ctx context.Context) {
	d := registry.Configuration().AutoSyncInterval
	if d == 0 || c.Cache == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
			c.Check()
		}
	}
}

func (c *ConflictChecker) Check() {
	caches, ok := c.Cache.(Cache)
	if !ok {
		return
	}

	var arr []*discovery.KeyValue
	for _, cache := range caches {
		cache.GetAll(&arr)
	}

	exists := make(map[string]*discovery.KeyValue)
	for _, v := range arr {
		key := util.BytesToStringWithNoCopy(v.Key)
		if kv, ok := exists[key]; ok {
			c.ConflictHandleFunc(kv, v)
			continue
		}
		exists[key] = v
	}
}

func NewConflictChecker(cache discovery.CacheReader, f func(origin, conflict *discovery.KeyValue)) *ConflictChecker {
	checker := &ConflictChecker{Cache: cache, ConflictHandleFunc: f}
	gopool.Go(checker.Run)
	return checker
}
