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
package sd

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"

	"context"
)

// Indexer searches k-v data.
// An indexer may search k-v data from many data sources,
// e.g. cache, cache, file, other Indexers...
type Indexer interface {
	// Search searches k-v data based on the input options
	Search(ctx context.Context, opts ...client.PluginOpOption) (*Response, error)
	// Creditable judges whether Indexer's search results are creditable
	// It is recommended to use cache only and not to call the cache
	// directly, If Indexer is not creditable.
	Creditable() bool
}
