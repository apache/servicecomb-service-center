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
	"github.com/apache/incubator-servicecomb-service-center/pkg/validate"
	"reflect"
)

const CAP_SIZE = 10

var handlersMap map[string][]Handler = make(map[string][]Handler, CAP_SIZE)

type Handler interface {
	Handle(i *Invocation)
}

func RegisterHandler(catalog string, h Handler) {
	handlers, ok := handlersMap[catalog]
	if !ok {
		handlers = make([]Handler, 0, CAP_SIZE)
	}
	handlers = append(handlers, h)
	handlersMap[catalog] = handlers

	t := validate.LoadStruct(reflect.ValueOf(h).Elem().Interface())
	util.Logger().Infof("register handler[%s] %s/%s", catalog, t.Type.PkgPath(), t.Type.Name())
}

func Handlers(catalog string) []Handler {
	return handlersMap[catalog]
}
