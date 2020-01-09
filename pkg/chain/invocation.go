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
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
)

type InvocationOption func(op InvocationOp) InvocationOp

type InvocationOp struct {
	Func  CallbackFunc
	Async bool
}

func WithFunc(f func(r Result)) InvocationOption {
	return func(op InvocationOp) InvocationOp { op.Func = f; return op }
}
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

func (i *Invocation) WithContext(key string, val interface{}) *Invocation {
	i.context.SetKV(key, val)
	return i
}

// Next is the method to go next step in handler chain
// WithFunc and WithAsyncFunc options can add customize callbacks in chain
// and the callbacks seq like below
// i.Success/Fail() -> CB1 ---> CB3 ----------> END           goroutine 0
//                          \-> CB2(async) \                  goroutine 1
//                                          \-> CB4(async)    goroutine 1 or 2
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
		cb(r)
		callback(f, async, r)
	}
}

func callback(f CallbackFunc, async bool, r Result) {
	c := Callback{Func: f, Async: async}
	c.Invoke(r)
}

func (i *Invocation) Invoke(f CallbackFunc) {
	defer func() {
		itf := recover()
		if itf == nil {
			return
		}
		log.LogPanic(itf)

		i.Fail(errorsEx.RaiseError(itf))
	}()
	i.Func = f
	i.chain.Next(i)
}

func NewInvocation(ctx context.Context, ch Chain) (inv Invocation) {
	inv.Init(ctx, ch)
	return inv
}
