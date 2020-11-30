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

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

var pool = gopool.New(context.Background())

type CallbackFunc func(r Result)

type Result struct {
	OK   bool
	Err  error
	Args []interface{}
}

func (r Result) String() string {
	if r.OK {
		return "OK"
	}
	return r.Err.Error()
}

type Callback struct {
	Func  CallbackFunc
	Async bool
}

func (cb *Callback) Invoke(r Result) {
	if cb.Async {
		pool.Do(func(_ context.Context) {
			cb.Func(r)
		})
		return
	}
	defer log.Recover()
	cb.Func(r)
}

func (cb *Callback) Fail(err error, args ...interface{}) {
	cb.Invoke(Result{
		OK:   false,
		Err:  err,
		Args: args,
	})
}

func (cb *Callback) Success(args ...interface{}) {
	cb.Invoke(Result{
		OK:   true,
		Args: args,
	})
}
