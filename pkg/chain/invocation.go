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
package chain

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
)

type Invocation struct {
	Callback
	context *util.StringContext
	chain   Chain
}

func (i *Invocation) Init(ctx context.Context, ch Chain) {
	i.context = util.NewStringContext(ctx)
	i.chain = ch
}

func (i *Invocation) Context() context.Context {
	return i.context
}

func (i *Invocation) WithContext(key string, val interface{}) *Invocation {
	i.context.SetKV(key, val)
	return i
}

func (i *Invocation) Next() {
	i.chain.Next(i)
}

func (i *Invocation) Invoke(f func(r Result)) {
	i.Func = f
	i.Next()
}

func NewInvocation(ctx context.Context, ch Chain) Invocation {
	var inv Invocation
	inv.Init(ctx, ch)
	return inv
}
