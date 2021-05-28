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

package buildin

import (
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"net/http"
	"strings"
)

var ErrCtxMatchPatternNotFound = errors.New("CtxMatchPattern not found")

var APIMapping = map[string]ParseFunc{}

type ParseFunc func(r *http.Request) (*auth.ResourceScope, error)

// ApplyAll work when no api registered by RegisterParseFunc matched
func ApplyAll(r *http.Request) (*auth.ResourceScope, error) {
	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		log.Error("CtxMatchPattern not found", nil)
		return nil, ErrCtxMatchPatternNotFound
	}
	return &auth.ResourceScope{
		Type: rbacmodel.GetResource(apiPath),
		Verb: rbac.MethodToVerbs[r.Method],
	}, nil
}

func FromRequest(r *http.Request) *auth.ResourceScope {
	apiPath, ok := r.Context().Value(rest.CtxMatchPattern).(string)
	if !ok {
		log.Error("CtxMatchPattern not found", nil)
		return nil
	}

	resource, err := GetAPIParseFunc(apiPath)(r)
	if err != nil {
		log.Error(fmt.Sprintf("parse from request failed"), err)
	}
	return resource
}

func GetAPIParseFunc(apiPattern string) ParseFunc {
	var (
		pf      ParseFunc = ApplyAll
		matched string
	)
	for pattern, f := range APIMapping {
		if apiPattern == pattern {
			return f
		}
		if len(matched) < len(pattern) && strings.Index(apiPattern, pattern) == 0 {
			pf = f
			matched = pattern
		}
	}
	return pf
}

func RegisterParseFunc(apiPathPrefix string, f ParseFunc) {
	APIMapping[apiPathPrefix] = f
}
