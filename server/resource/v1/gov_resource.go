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
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	"github.com/apache/servicecomb-service-center/server/service/gov"
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
)

//Create gov config
func (t *Governance) Create(w http.ResponseWriter, req *http.Request) {
	kind := req.URL.Query().Get(KindKey)
	project := req.URL.Query().Get(ProjectKey)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	err = gov.Create(kind, project, body)
	if err != nil {
		log.Error("create gov err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

//Put gov config
func (t *Governance) Put(w http.ResponseWriter, req *http.Request) {
	kind := req.URL.Query().Get(KindKey)
	id := req.URL.Query().Get(IDKey)
	project := req.URL.Query().Get(ProjectKey)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("read body err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	err = gov.Update(id, kind, project, body)
	if err != nil {
		log.Error("create gov err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

//List return all gov config
func (t *Governance) List(w http.ResponseWriter, req *http.Request) {
	kind := req.URL.Query().Get(KindKey)
	project := req.URL.Query().Get(ProjectKey)
	app := req.URL.Query().Get(AppKey)
	environment := req.URL.Query().Get(EnvironmentKey)
	body, err := gov.List(kind, project, app, environment)
	if err != nil {
		log.Error("create gov err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	_, err = w.Write(body)
	if err != nil {
		log.Error("", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
}

//Get gov config
func (t *Governance) Get(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(IDKey)
	project := req.URL.Query().Get(ProjectKey)
	body, err := gov.Get(id, project)
	if err != nil {
		log.Error("create gov err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	_, err = w.Write(body)
	if err != nil {
		log.Error("", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set(rest.HeaderContentType, rest.ContentTypeJSON)
}

//Delete delete gov config
func (t *Governance) Delete(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(IDKey)
	project := req.URL.Query().Get(ProjectKey)
	err := gov.Delete(id, project)
	if err != nil {
		log.Error("create gov err", err)
		controller.WriteError(w, discovery.ErrInternal, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (t *Governance) URLPatterns() []rest.Route {
	return []rest.Route{
		//servicecomb.marker.{name}
		//servicecomb.rateLimiter.{name}
		//....
		{Method: http.MethodPost, Path: "/v1/:project/gov/" + KindKey, Func: t.Create},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + KindKey, Func: t.List},
		{Method: http.MethodGet, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Get},
		{Method: http.MethodPut, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Put},
		{Method: http.MethodDelete, Path: "/v1/:project/gov/" + KindKey + "/" + IDKey, Func: t.Delete},
	}
}
