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
package domain

import (
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	"net/http"
)

const (
	DEFAULT_PROJECT = "default"
)

var NO_CHECK_URL = map[string]bool{"/version": true, "/health": true}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	request := r
	tenant := ""
	project := ""
	ctx := r.Context()
	tenant, project = GetTenantProjectFromHeader(r)
	if len(tenant) == 0 || len(project) == 0 {
		err := errors.New("Header does not contain domain.")

		util.Logger().Errorf(err, "Invalid Request URI %s", r.RequestURI)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(util.StringToBytesWithNoCopy(err.Error()))
		return err
	}

	ctx = util.NewContext(ctx, "tenant", tenant)
	ctx = util.NewContext(ctx, "project", project)
	request = r.WithContext(ctx)
	*r = *request
	return nil
}

func IsSkip(uri string) bool {
	if b, ok := NO_CHECK_URL[uri]; ok {
		return b
	}
	return false
}

func GetTenantProjectFromHeader(r *http.Request) (string, string) {
	var domain, project string
	domain = r.Header.Get("X-Tenant-Name")
	if len(domain) == 0 {
		domain = r.Header.Get("x-domain-name")
		if len(domain) == 0 {
			if IsSkip(r.RequestURI) {
				return core.REGISTRY_TENANT, core.REGISTRY_PROJECT
			}
			return "", ""
		}
	}
	project = r.Header.Get("X-Project-Name")
	if len(project) == 0 {
		project = DEFAULT_PROJECT
	}
	return domain, project
}
