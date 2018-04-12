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
)

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
	Func  func(r Result)
	Async bool
}

func (cb *Callback) Invoke(r Result) {
	if cb.Async {
		go syncInvoke(cb.Func, r)
		return
	}
	syncInvoke(cb.Func, r)
}

func syncInvoke(f func(r Result), r Result) {
	defer util.RecoverAndReport()
	if f == nil {
		util.Logger().Errorf(nil, "Callback function is nil. result: %s,", r)
		return
	}
	f(r)
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
