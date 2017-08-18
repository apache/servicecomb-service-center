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
	"github.com/ServiceComb/service-center/server/helper"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/errors"
	"net/http"
)

const (
	TIME_FORMAT = "2006-01-02T15:04:05Z07:00"
)

func Intercept(w http.ResponseWriter, r *http.Request) error {
	util.LOGGER.Debugf("Intercept Domain")

	request := r
	tenant := ""
	project := ""
	var err error
	ctx := r.Context()
	tenant, project, err = helper.GetTenantProjectFromHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(util.StringToBytesWithNoCopy(err.Error()))
		return err
	}
	if len(tenant) == 0 || len(project) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(util.StringToBytesWithNoCopy("Domain or project from token is empty."))
		return errors.New("Domain or project from token is empty.")
	}

	ctx = util.NewContext(ctx, "tenant", tenant)
	ctx = util.NewContext(ctx, "project", project)
	request = r.WithContext(ctx)
	*r = *request
	return nil
}
