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

package state

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
)

// State is used to do service discovery.
// To improve the performance, State may use cache firstly in
// service pkg.
type State interface {
	kvstore.Runnable
	// Indexer is used to search data from the cache.
	kvstore.Indexer
	// Cacher is used to manage the registry's cache.
	kvstore.Cacher
}

// Repository creates States
type Repository interface {
	// New news an instance of specify Type repo
	New(t kvstore.Type, cfg *kvstore.Options) State
}
