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
	"errors"
)

const CAP_SIZE = 10

type Invocation struct {
	Callback
	handlerContext map[string]interface{}
	context        map[string]interface{}
	chain          *Chain
}

func (i *Invocation) Init(ch *Chain) {
	i.handlerContext = make(map[string]interface{}, CAP_SIZE)
	i.context = make(map[string]interface{}, CAP_SIZE)
	i.chain = ch
}

func (i *Invocation) HandlerContext() map[string]interface{} {
	return i.handlerContext
}

func (i *Invocation) Context() map[string]interface{} {
	return i.context
}

func (i *Invocation) WithHandlerContext(key string, val interface{}) *Invocation {
	i.handlerContext[key] = val
	return i
}

func (i *Invocation) WithContext(key string, val interface{}) *Invocation {
	i.context[key] = val
	return i
}

func (i *Invocation) Next() {
	if i.chain == nil {
		i.Fail(errors.New("Can not find any chain for this invocation"))
		return
	}
	i.chain.Next(i)
}

func (i *Invocation) Invoke(f func(r Result)) {
	i.Func = f
	i.Next()
}

func NewInvocation(ch *Chain) *Invocation {
	var inv Invocation
	inv.Init(ch)
	return &inv
}
