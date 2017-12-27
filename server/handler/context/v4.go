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
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"net/http"
	"net/url"
	"strings"
)

type v4Context struct {
}

func (v *v4Context) IsMatch(r *http.Request) bool {
	return strings.Index(r.RequestURI, "/v4/") == 0
}

func (v *v4Context) Do(r *http.Request) error {
	ctx := r.Context()

	if len(util.ParseProject(ctx)) == 0 {
		path, err := url.PathUnescape(r.RequestURI)
		if err != nil {
			util.Logger().Errorf(err, "Invalid Request URI %s", r.RequestURI)
			return err
		}

		util.SetRequestContext(r, "project", v.parseProjectFromPath(path))
	}

	if len(util.ParseDomain(ctx)) == 0 {
		domain := r.Header.Get("X-Domain-Name")
		if len(domain) == 0 {
			err := errors.New("Header does not contain domain.")
			util.Logger().Errorf(err, "Invalid Request URI %s", r.RequestURI)
			return err
		}
		util.SetRequestContext(r, "domain", domain)
	}
	return nil
}

func (v *v4Context) parseProjectFromPath(path string) string {
	start := len("/v4/")
	end := start + strings.Index(path[start:], "/")

	project := strings.TrimSpace(path[start:end])
	if len(project) == 0 {
		project = core.REGISTRY_PROJECT
	}
	return project
}
