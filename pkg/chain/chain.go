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
	errorsEx "github.com/ServiceComb/service-center/pkg/errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/pkg/validate"
	"reflect"
)

type Chain struct {
	name         string
	handlers     []Handler
	currentIndex int
}

func (c *Chain) Init(chainName string, hs []Handler) {
	c.name = chainName
	c.currentIndex = -1
	if len(hs) > 0 {
		c.handlers = make([]Handler, len(hs))
		copy(c.handlers, hs)
	}
}

func (c *Chain) Name() string {
	return c.name
}

func (c *Chain) syncNext(i *Invocation) {
	defer func() {
		itf := recover()
		if itf == nil {
			return
		}
		t := validate.LoadStruct(reflect.ValueOf(c.handlers[c.currentIndex]).Elem().Interface())
		util.Logger().Errorf(nil, "recover from '%s/%s.Handle()'! %v", t.Type.PkgPath(), t.Type.Name(), itf)

		i.Fail(errorsEx.RaiseError(itf))
	}()

	if c.currentIndex >= len(c.handlers)-1 {
		i.Success()
		return
	}
	c.currentIndex += 1
	c.handlers[c.currentIndex].Handle(i)
}

func (c *Chain) Next(i *Invocation) {
	go c.syncNext(i)
}

func NewChain(name string, handlers []Handler) Chain {
	var ch Chain
	ch.Init(name, handlers)
	return ch
}
