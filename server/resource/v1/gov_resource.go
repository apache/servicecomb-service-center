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

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/server/common"
	"io/ioutil"
	"net/http"

	model "github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/apache/servicecomb-service-center/server/service/gov/kie"
	"github.com/go-chassis/cari/discovery"
)

type Governance struct {
}

func getCtx(r *http.Request) (context.Context, error) {
	ctx := context.TODO()
	query := r.URL.Query()
	if query.Get(common.KindKey) != "" {
		context.WithValue(ctx, common.KindKey, query.Get(common.KindKey))
	}
	if query.Get(common.ProjectKey) != "" {
		context.WithValue(ctx, common.ProjectKey, query.Get(common.ProjectKey))
	}
	if query.Get(common.IDKey) != "" {
		context.WithValue(ctx, common.IDKey, query.Get(common.IDKey))
	}
	if query.Get(common.AppKey) != "" {
		context.WithValue(ctx, common.AppKey, query.Get(common.AppKey))
	}
	if query.Get(common.EnvironmentKey) != "" {
		context.WithValue(ctx, common.EnvironmentKey, query.Get(common.EnvironmentKey))
	}
	if r.Body != nil {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("read body err", err)
			return nil, err
		}
		context.WithValue(ctx, common.SpecKey, body)
	}

	h := r.Header
	if h[common.HeaderAuth] != nil {
		context.WithValue(ctx, common.HeaderAuth, h[common.HeaderAuth])
	}
	return ctx, nil
}

//Create gov config
func (t *Governance) Create(w http.ResponseWriter, r *http.Request) {
	ctx, err := getCtx(r)
	if err != nil {
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	p := &model.Policy{
		GovernancePolicy: &model.GovernancePolicy{Selector: &model.Selector{}},
	}

	body := ctx.Value(common.SpecKey).([]byte)
	err = json.Unmarshal(body, p)
	if err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}

	kind := ctx.Value(common.KindKey).(string)
	err = gov.ValidateSpec(kind, p.Spec)
	if err != nil {
		log.Error("validate policy err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}

	id, err := gov.Create(ctx, p)
	if err != nil {
		if _, ok := err.(*kie.ErrIllegalItem); ok {
			log.Error("", err)
			rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
			return
		}
		processError(w, err, "create gov data err")
		return
	}
	log.Info(fmt.Sprintf("created %+v", p))
	rest.WriteResponse(w, r, nil, &model.Policy{GovernancePolicy: &model.GovernancePolicy{ID: string(id)}})
}

//Put gov config
func (t *Governance) Put(w http.ResponseWriter, r *http.Request) {
	ctx, err := getCtx(r)
	if err != nil {
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	p := &model.Policy{}

	body := ctx.Value(common.SpecKey).([]byte)
	err = json.Unmarshal(body, p)
	if err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}

	kind := ctx.Value(common.KindKey).(string)
	err = gov.ValidateSpec(kind, p.Spec)
	if err != nil {
		log.Error("validate policy err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	log.Info(fmt.Sprintf("update %v", &p))
	err = gov.Update(ctx, p)
	if err != nil {
		if _, ok := err.(*kie.ErrIllegalItem); ok {
			log.Error("", err)
			rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
			return
		}
		processError(w, err, "put gov err")
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

//ListOrDisPlay return all gov config
func (t *Governance) ListOrDisPlay(w http.ResponseWriter, r *http.Request) {
	ctx, _ := getCtx(r)
	var body []byte
	var err error

	if ctx.Value(common.KindKey).(string) == common.DisplayKey {
		body, err = gov.Display(ctx)
	} else {
		body, err = gov.List(ctx)
	}
	if err != nil {
		processError(w, err, "list gov err")
		return
	}
	rest.WriteResponse(w, r, nil, body)
}

//Get gov config
func (t *Governance) Get(w http.ResponseWriter, r *http.Request) {
	ctx, _ := getCtx(r)
	body, err := gov.Get(ctx)
	if err != nil {
		processError(w, err, "get gov err")
		return
	}
	rest.WriteResponse(w, r, nil, body)
}

//Delete delete gov config
func (t *Governance) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, _ := getCtx(r)
	err := gov.Delete(ctx)
	if err != nil {
		processError(w, err, "delete gov err")
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func processError(w http.ResponseWriter, err error, msg string) {
	log.Error(msg, err)
	rest.WriteError(w, discovery.ErrInternal, err.Error())
}

func (t *Governance) URLPatterns() []rest.Route {
	return []rest.Route{
		//servicecomb.marker.{name}
		//servicecomb.rateLimiter.{name}
		//....
		{Method: http.MethodPost, Path: "/v1/:project/gov/" + common.KindKey, Func: t.Create},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + common.KindKey, Func: t.ListOrDisPlay},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + common.KindKey + "/" + common.IDKey, Func: t.Get},
		{Method: http.MethodPut, Path: "/v1/:project/gov/" + common.KindKey + "/" + common.IDKey, Func: t.Put},
		{Method: http.MethodDelete, Path: "/v1/:project/gov/" + common.KindKey + "/" + common.IDKey, Func: t.Delete},
	}
}
