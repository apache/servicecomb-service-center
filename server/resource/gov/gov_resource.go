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

package gov

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	model "github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/service/grc"
	"github.com/go-chassis/cari/discovery"
)

type Governance struct {
}

const (
	AppKey         = "app"
	EnvironmentKey = "environment"
	KindKey        = ":kind"
	ProjectKey     = ":project"
	IDKey          = ":id"
	DisplayKey     = "display"
)

//Create gov config
func (t *Governance) Create(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	kind := query.Get(KindKey)
	project := query.Get(ProjectKey)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("read body err", err)
		rest.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	p := &model.Policy{
		GovernancePolicy: &model.GovernancePolicy{Selector: model.Selector{}},
	}
	err = json.Unmarshal(body, p)
	if err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	err = grc.ValidatePolicySpec(kind, p.Spec)
	if err != nil {
		log.Error(fmt.Sprintf("validate policy [%s] err", kind), err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	id, err := grc.Create(r.Context(), kind, project, p)
	if err != nil {
		if _, ok := err.(*grc.ErrIllegalItem); ok {
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
	query := r.URL.Query()
	kind := query.Get(KindKey)
	id := query.Get(IDKey)
	project := query.Get(ProjectKey)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		processError(w, err, "read body err")
		return
	}
	p := &model.Policy{}
	err = json.Unmarshal(body, p)
	if err != nil {
		log.Error("json err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	err = grc.ValidatePolicySpec(kind, p.Spec)
	if err != nil {
		log.Error("validate policy err", err)
		rest.WriteError(w, discovery.ErrInvalidParams, err.Error())
		return
	}
	log.Info(fmt.Sprintf("update %v", &p))
	err = grc.Update(r.Context(), kind, id, project, p)
	if err != nil {
		if _, ok := err.(*grc.ErrIllegalItem); ok {
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
	query := r.URL.Query()
	kind := query.Get(KindKey)
	project := query.Get(ProjectKey)
	app := query.Get(AppKey)
	environment := query.Get(EnvironmentKey)
	var body []byte
	var err error
	if kind == DisplayKey {
		body, err = grc.Display(r.Context(), project, app, environment)
	} else {
		body, err = grc.List(r.Context(), kind, project, app, environment)
	}
	if err != nil {
		processError(w, err, "list gov err")
		return
	}
	rest.WriteResponse(w, r, nil, body)
}

//Get gov config
func (t *Governance) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	kind := query.Get(KindKey)
	id := query.Get(IDKey)
	project := query.Get(ProjectKey)
	body, err := grc.Get(r.Context(), kind, id, project)
	if err != nil {
		processError(w, err, "get gov err")
		return
	}
	rest.WriteResponse(w, r, nil, body)
}

//Delete delete gov config
func (t *Governance) Delete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	kind := query.Get(KindKey)
	id := query.Get(IDKey)
	project := query.Get(ProjectKey)
	err := grc.Delete(r.Context(), kind, id, project)
	if err != nil {
		processError(w, err, "delete gov err")
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func processError(w http.ResponseWriter, err error, msg string) {
	log.Error(msg, err)
	rest.WriteServiceError(w, err)
}

func (t *Governance) URLPatterns() []rest.Route {
	return []rest.Route{
		//servicecomb.marker.{name}
		//servicecomb.rateLimiter.{name}
		//....
		{Method: http.MethodPost, Path: "/v1/:project/gov/" + KindKey, Func: t.Create},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + KindKey, Func: t.ListOrDisPlay},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Get},
		{Method: http.MethodPut, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Put},
		{Method: http.MethodDelete, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Delete},
	}
}
