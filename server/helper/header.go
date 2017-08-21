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
package helper

import (
	"errors"
	"fmt"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/util"
	"net/http"
	"strings"
)

const (
	DEFAULT_PROJECT = "default"
)

var NO_CHEACK_URL = map[string]bool{"/version": true, "/health": true}

func GetTenantProjectFromHeader(r *http.Request) (string, string, error) {
	var domain, project string
	domain = r.Header.Get("X-Tenant-Name")
	if len(domain) == 0 {
		domain = r.Header.Get("x-domain-name")
		if len(domain) == 0 {
			if _, ok := NO_CHEACK_URL[r.RequestURI]; ok {
				return core.REGISTRY_TENANT, core.REGISTRY_PROJECT, nil
			}
			util.LOGGER.Errorf(nil, "%s does not contain domain.", r.RequestURI)
			return "", "", errors.New(fmt.Sprintf("Header does not contain tenant.Invalid Request URI %s", r.RequestURI))
		}
	}
	project = r.Header.Get("X-Project-Name")
	if len(project) == 0 {
		project = DEFAULT_PROJECT
	}
	return domain, project, nil
}

func GetRealIP(r *http.Request) string {
	addrs := strings.Split(r.RemoteAddr, ":")
	if len(addrs) > 0 {
		return addrs[0]
	}
	return ""
}

func addIPToContext(r *http.Request) {
	terminalIP := GetRealIP(r)
	ctx := r.Context()
	ctx = util.NewContext(ctx, "x-remote-ip", terminalIP)
	request := r.WithContext(ctx)
	*r = *request
}

func InitContext(r *http.Request) {
	addIPToContext(r)
}
