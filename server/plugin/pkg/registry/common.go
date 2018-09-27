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

package registry

import (
	"golang.org/x/net/context"
	"time"
)

const (
	Get ActionType = iota
	Put
	Delete
)

const (
	SORT_NONE SortOrder = iota
	SORT_ASCEND
	SORT_DESCEND
)

const (
	CMP_VERSION CompareType = iota
	CMP_CREATE
	CMP_MOD
	CMP_VALUE
)

const (
	CMP_EQUAL CompareResult = iota
	CMP_GREATER
	CMP_LESS
	CMP_NOT_EQUAL
)

const (
	MODE_BOTH CacheMode = iota
	MODE_CACHE
	MODE_NO_CACHE
)

const (
	// grpc does not allow to transport a large body more then 4MB in a request
	DEFAULT_PAGE_COUNT = 4096
	// the timeout dial to etcd
	defaultDialTimeout    = 10 * time.Second
	defaultRequestTimeout = 30 * time.Second
)

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, defaultRegistryConfig.RequestTimeOut)
}
