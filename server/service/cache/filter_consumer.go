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

	"github.com/apache/servicecomb-service-center/pkg/cache"
)

type ConsumerFilter struct {
}

func (f *ConsumerFilter) Name(ctx context.Context, _ *cache.Node) string {
	return ctx.Value(CtxFindConsumer).(string)
}

func (f *ConsumerFilter) Init(ctx context.Context, parent *cache.Node) (node *cache.Node, err error) {
	node = cache.NewNode()
	node.Cache.Set(Dep, &DependencyRuleItem{})
	return
}
