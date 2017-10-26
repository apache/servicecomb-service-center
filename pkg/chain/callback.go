//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package chain

import (
	"github.com/ServiceComb/service-center/pkg/util"
)

type Result struct {
	OK   bool
	Err  error
	Args []interface{}
}

type CallbackFunc func(r Result)

type Callback struct {
	Func CallbackFunc
}

func (cb *Callback) Invoke(r Result) {
	go func() {
		defer util.RecoverAndReport()
		cb.Func(r)
	}()
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

func NewCallback(f CallbackFunc) Callback {
	return Callback{
		Func: f,
	}
}
