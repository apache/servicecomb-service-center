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
package v4

import (
	"errors"
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	"net/http"
	"net/url"
	"strings"
)

var router http.Handler

func init() {
	util.Logger().Debugf("init router")
	router = roa.InitROAServerHandler()
	initRouter()
}

func initRouter() {
	roa.RegisterFilter(&v4Context{})
	roa.RegisterServent(&MainService{})
	roa.RegisterServent(&MicroServiceService{})
	roa.RegisterServent(&TagService{})
	roa.RegisterServent(&RuleService{})
	roa.RegisterServent(&MicroServiceInstanceService{})
	roa.RegisterServent(&WatchService{})
	roa.RegisterServent(&GovernService{})
}

//GetRouter return the router fo REST service
func GetRouter() http.Handler {
	return router
}

type v4Context struct {
}

func (v *v4Context) IsMatch(r *http.Request) bool {
	return strings.Index(r.RequestURI, "/v4/") == 0
}

func (v *v4Context) Do(r *http.Request) error {
	ctx := r.Context()
	if ctx.Value("tenant") == nil {
		path, err := url.PathUnescape(r.RequestURI)
		if err != nil {
			util.Logger().Errorf(err, "Invalid Request URI %s", r.RequestURI)
			return err
		}

		start := len("/v4/")
		end := start + strings.Index(path[start:], "/")

		tenant := strings.TrimSpace(path[start:end])
		if len(tenant) == 0 {
			err := errors.New("Header does not contain domain.")
			util.Logger().Errorf(err, "Invalid Request URI %s", r.RequestURI)
			return err
		}
		util.SetReqCtx(r, "tenant", tenant)
	}

	if ctx.Value("project") == nil {
		project := r.Header.Get("X-Domain-Name")
		if len(project) == 0 {
			project = core.REGISTRY_PROJECT
		}
		util.SetReqCtx(r, "project", project)
	}
	return nil
}
