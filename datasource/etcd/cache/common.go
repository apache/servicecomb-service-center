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

package cache

import (
	"context"
	"sort"
	"sync"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type ClustersIndex map[string]int

const (
	CtxConsumerID          util.CtxKey = "_consumer"
	CtxProviderKey         util.CtxKey = "_provider"
	CtxProviderInstanceKey util.CtxKey = "_provider_instance"
	CtxTags                util.CtxKey = "_tags"
	CtxRequestRev          util.CtxKey = "_rev"

	FindResult = "_find"
	DepResult  = "_dep"

	DefaultCacheMaxSize = 10000
)

var (
	buildClustersIndexOnce sync.Once
	clustersIndex          = make(ClustersIndex)
)

func getOrCreateClustersIndex() ClustersIndex {
	buildClustersIndexOnce.Do(func() {
		var clusters []string
		resp, _ := datasource.GetSCManager().GetClusters(context.Background())
		for name := range resp {
			clusters = append(clusters, name)
		}
		sort.Strings(clusters)
		for i, name := range clusters {
			clustersIndex[name] = i
		}
	})
	return clustersIndex
}
