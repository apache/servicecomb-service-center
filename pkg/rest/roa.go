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
package rest

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"net/http"
	"reflect"
)

var (
	serverHandler *ROAServerHandler
)

func init() {
	initROAServerHandler()
}

type ROAServantService interface {
	URLPatterns() []Route
}

func initROAServerHandler() *ROAServerHandler {
	serverHandler = NewROAServerHander()
	return serverHandler
}

// servant must be an pointer to service object
func RegisterServent(servant interface{}) {
	val := reflect.ValueOf(servant)
	ind := reflect.Indirect(val)
	typ := ind.Type()
	name := typ.PkgPath() + "." + typ.Name()
	if val.Kind() != reflect.Ptr {
		util.Logger().Errorf(nil, "<rest.RegisterServent> cannot use non-ptr servant struct `%s`", name)
		return
	}

	urlPatternFunc := val.MethodByName("URLPatterns")
	if !urlPatternFunc.IsValid() {
		util.Logger().Errorf(nil, "<rest.RegisterServent> no 'URLPatterns' function in servant struct `%s`", name)
		return
	}

	vals := urlPatternFunc.Call([]reflect.Value{})
	if len(vals) <= 0 {
		util.Logger().Errorf(nil, "<rest.RegisterServent> call 'URLPatterns' function failed in servant struct `%s`", name)
		return
	}

	val0 := vals[0]
	if !val.CanInterface() {
		util.Logger().Errorf(nil, "<rest.RegisterServent> result of 'URLPatterns' function not interface type in servant struct `%s`", name)
		return
	}

	if routes, ok := val0.Interface().([]Route); ok {
		util.Logger().Infof("register servant %s", name)
		for _, route := range routes {
			err := serverHandler.addRoute(&route)
			if err != nil {
				util.Logger().Errorf(err, "register route failed.")
			}
		}
	} else {
		util.Logger().Errorf(nil, "<rest.RegisterServent> result of 'URLPatterns' function not []*Route type in servant struct `%s`", name)
	}
}

//GetRouter return the router fo REST service
func GetRouter() http.Handler {
	return serverHandler
}
