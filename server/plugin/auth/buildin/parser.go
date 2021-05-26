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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/plugin/auth"
	"github.com/apache/servicecomb-service-center/server/service/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"net/http"
	"strings"
)

var (
	//TODO ...
	APIAccountList = "/v4/accounts"
	APIRoleList    = "/v4/roles"
	APIOps         = "/v4/:project/admin"
	APIGov         = "/v1/:project/gov"
)

var APIMapping = map[string]ParseFunc{}

type ParseFunc func(r *http.Request) ([]map[string]string, error)

func ApplyAll(_ *http.Request) ([]map[string]string, error) {
	return nil, nil
}

func FromRequest(r *http.Request) *auth.ResourceScope {
	apiPath := r.Context().Value(rest.CtxMatchPattern).(string)

	resource := rbacmodel.GetResource(apiPath)
	labels, err := GetAPIParseFunc(apiPath)(r)
	if err != nil {
		log.Error(fmt.Sprintf("parse from request failed"), err)
		return nil
	}
	return &auth.ResourceScope{
		Type:   resource,
		Labels: labels,
		Verb:   rbac.MethodToVerbs[r.Method],
	}
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
