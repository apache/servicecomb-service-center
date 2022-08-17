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
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

type InvocationOption func(op InvocationOp) InvocationOp

type InvocationOp struct {
	Func  CallbackFunc
	Async bool
}

// WithFunc append a func to the begin of invocation callback list
func WithFunc(f func(r Result)) InvocationOption {
	return func(op InvocationOp) InvocationOp { op.Func = f; return op }
}

// WithAsyncFunc called concurrently after all WithFunc finish
func WithAsyncFunc(f func(r Result)) InvocationOption {
	return func(op InvocationOp) InvocationOp { op.Func = f; op.Async = true; return op }
}

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

func (i *Invocation) WithContext(key util.CtxKey, val interface{}) *Invocation {
	i.context.SetKV(key, val)
	return i
}

// Next is the method to go next step in handler chain
// WithFunc and WithAsyncFunc options can add customize callbacks in chain
// and the callbacks seq like below:
// when i.Next(WithFunc(CB1)).Next(WithAsyncFunc(CB2)).Next(WithFunc(CB3)).Invoke(CB0)
// then i.Success/Fail() -> CB3 -> CB1 -> CB0(invoke)             goroutine 0
//
//	\-> CB2(async)    goroutine 1
func (i *Invocation) Next(opts ...InvocationOption) {
	var op InvocationOp
	for _, opt := range opts {
		op = opt(op)
	}

	i.setCallback(op.Func, op.Async)
	i.chain.Next(i)
}

func (i *Invocation) setCallback(f CallbackFunc, async bool) {
	if f == nil {
		return
	}

	if i.Func == nil {
		i.Func = f
		i.Async = async
		return
	}
	cb := i.Func
	i.Func = func(r Result) {
		callback(cb, f, async, r)
	}
}

func callback(prev, next CallbackFunc, async bool, r Result) {
	// we make sure the all sync funcs called before the async funcs
	if async {
		prev(r)
	}

	c := Callback{Func: next, Async: async}
	c.Invoke(r)

	if !async {
		prev(r)
	}
}

func (i *Invocation) Invoke(last CallbackFunc) {
	defer func() {
		itf := recover()
		if itf == nil {
			return
		}
		log.Panic(itf)

		// this recover only catch the exceptions raised in sync invocations.
		// The async invocations will be catch by gopool pkg then it never
		// change the callback results.
		i.Fail(discovery.NewError(discovery.ErrInternal, fmt.Sprintf("%v", itf)))
	}()
	i.Func = last
	i.chain.Next(i)
}

func NewInvocation(ctx context.Context, ch Chain) (inv Invocation) {
	inv.Init(ctx, ch)
	return inv
}
