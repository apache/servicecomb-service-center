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

package context

import (
	"net/http"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
)

type v3Context struct {
}

func (v *v3Context) IsMatch(r *http.Request) bool {
	return strings.Index(r.RequestURI, "/registry/v3/") == 0
}

func (v *v3Context) Do(r *http.Request) error {
	ctx := r.Context()

	domain, project := util.ParseDomain(ctx), util.ParseProject(ctx)

	if len(domain) == 0 {
		domain = r.Header.Get("X-Tenant-Name")
		if len(domain) == 0 {
			domain = r.Header.Get("X-Domain-Name")
		}

		if len(domain) == 0 {
			domain = "default"
		}
		util.SetRequestContext(r, util.CtxDomain, domain)
	}

	if len(project) == 0 {
		util.SetRequestContext(r, util.CtxProject, core.RegistryProject)
	}

	return nil
}
