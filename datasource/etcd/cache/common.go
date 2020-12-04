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
	CtxFindConsumer         util.CtxKey = "_consumer"
	CtxFindProvider         util.CtxKey = "_provider"
	CtxFindProviderInstance util.CtxKey = "_provider_instance"
	CtxFindTags             util.CtxKey = "_tags"
	CtxFindRequestRev       util.CtxKey = "_rev"

	Find = "_find"
	Dep  = "_dep"
)

var (
	buildClustersIndexOnce sync.Once
	clustersIndex          = make(ClustersIndex)
)

func getOrCreateClustersIndex() ClustersIndex {
	buildClustersIndexOnce.Do(func() {
		var clusters []string
		resp, _ := datasource.Instance().GetClusters(context.Background())
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
