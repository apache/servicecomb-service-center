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

package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type v4Context struct {
}

func (v *v4Context) IsMatch(r *http.Request) bool {
	return strings.Index(r.RequestURI, "/v4/") == 0
}

func (v *v4Context) Write(c *fiber.Ctx) {
	domain, project := util.ParseDomain(c.UserContext()), util.ParseProject(c.UserContext())

	if len(domain) == 0 {
		domain = c.Get("X-Domain-Name")
		if len(domain) == 0 {
			domain = "default"
		}
		util.SetFiberContext(c, util.CtxDomain, domain)
	}

	if len(project) == 0 {
		project = c.Params("project")
		if len(project) == 0 {
			project = datasource.RegistryProject
		}
		util.SetFiberContext(c, util.CtxProject, project)
	}
}
