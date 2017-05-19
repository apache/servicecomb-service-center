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
	"github.com/servicecomb/service-center/util"
	"github.com/servicecomb/service-center/util/errors"
	"golang.org/x/net/context"
	"net/http"
	"github.com/servicecomb/service-center/server/helper"
)

const (
	TIME_FORMAT = "2006-01-02T15:04:05Z07:00"
)

func Intercept(w http.ResponseWriter, r *http.Request) error {
	util.LOGGER.Info("Intercept Domain")

	request := r
	tenant := ""
	project := ""
	var err error
	ctx := r.Context()
	tenant, project, err = helper.GetTenantProjectFromHeader(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return err
	}
	if len(tenant) == 0 || len(project) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errors.New("Domain or project from token is empty.").Error()))
		return errors.New("Domain or project from token is empty.")
	}

	ctx = context.WithValue(ctx, "tenant", tenant)
	ctx = context.WithValue(ctx, "project", project)
	request = r.WithContext(ctx)
	*r = *request
	return nil
}