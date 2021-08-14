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

package broker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/broker/brokerpb"
	"github.com/apache/servicecomb-service-center/server/rest/controller"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

const DefaultScheme = "http"

type Controller struct {
}

func (brokerService *Controller) URLPatterns() []rest.Route {
	return []rest.Route{
		// for handling broker requests
		{Method: http.MethodGet,
			Path: "/",
			Func: brokerService.GetHome},
		{Method: rest.HTTPMethodPut,
			Path: "/pacts/provider/:providerId/consumer/:consumerId/version/:number",
			Func: brokerService.PublishPact},
		{Method: http.MethodGet,
			Path: "/pacts/provider/:providerId/latest",
			Func: brokerService.GetAllProviderPacts},
		{Method: http.MethodGet,
			Path: "/pacts/provider/:providerId/consumer/:consumerId/version/:number",
			Func: brokerService.GetPactsOfProvider},
		{Method: rest.HTTPMethodDelete,
			Path: "/pacts/delete",
			Func: brokerService.DeletePacts},
		{Method: rest.HTTPMethodPost,
			Path: "/pacts/provider/:providerId/consumer/:consumerId/pact-version/:sha/verification-results",
			Func: brokerService.PublishVerificationResults},
		{Method: http.MethodGet,
			Path: "/verification-results/consumer/:consumerId/version/:consumerVersion/latest",
			Func: brokerService.RetrieveVerificationResults},
	}
}

func (brokerService *Controller) GetHome(w http.ResponseWriter, r *http.Request) {
	request := &brokerpb.BaseBrokerRequest{
		HostAddress: r.Host,
		Scheme:      getScheme(r),
	}
	resp, _ := ServiceAPI.GetBrokerHome(r.Context(), request)

	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (*Controller) PublishPact(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		PactLogger.Error("body err\n", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	request := &brokerpb.PublishPactRequest{
		ProviderId: query.Get(":providerId"),
		ConsumerId: query.Get(":consumerId"),
		Version:    query.Get(":number"),
		Pact:       message,
	}
	PactLogger.Infof("PublishPact: providerId = %s, consumerId = %s, version = %s\n",
		request.ProviderId, request.ConsumerId, request.Version)
	resp, err := ServiceAPI.PublishPact(r.Context(), request)
	if err != nil {
		log.Errorf(err, "can not push pact")
		controller.WriteError(w, scerr.ErrInternal, "can not push pact")
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (*Controller) GetAllProviderPacts(w http.ResponseWriter, r *http.Request) {
	request := &brokerpb.GetAllProviderPactsRequest{
		ProviderId: r.URL.Query().Get(":providerId"),
		BaseUrl: &brokerpb.BaseBrokerRequest{
			HostAddress: r.Host,
			Scheme:      getScheme(r),
		},
	}
	resp, err := ServiceAPI.GetAllProviderPacts(r.Context(), request /*, href*/)
	if err != nil {
		PactLogger.Errorf(err, "can not get pacts")
		controller.WriteError(w, scerr.ErrInternal, "can not get pacts")
		return
	}
	linksObj, err := json.Marshal(resp)
	if err != nil {
		PactLogger.Errorf(err, "invalid ProviderPacts")
		controller.WriteError(w, scerr.ErrInternal, "Marshal error")
		return
	}
	PactLogger.Infof("Pact info: %s\n", string(linksObj))
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (*Controller) GetPactsOfProvider(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	request := &brokerpb.GetProviderConsumerVersionPactRequest{
		ProviderId: query.Get(":providerId"),
		ConsumerId: query.Get(":consumerId"),
		Version:    query.Get(":number"),
		BaseUrl: &brokerpb.BaseBrokerRequest{
			HostAddress: r.Host,
			Scheme:      getScheme(r),
		},
	}

	resp, _ := ServiceAPI.GetPactsOfProvider(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	//controller.WriteResponse(w, respInternal, resp.Pact)
	controller.WriteJSONIfSuccess(w, respInternal, resp.Pact)
}

func (*Controller) DeletePacts(w http.ResponseWriter, r *http.Request) {
	resp, _ := ServiceAPI.DeletePacts(r.Context(), &brokerpb.BaseBrokerRequest{
		HostAddress: r.Host,
		Scheme:      getScheme(r),
	})
	controller.WriteResponse(w, r, resp, nil)
}

func (*Controller) PublishVerificationResults(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		PactLogger.Error("body err", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request := &brokerpb.PublishVerificationRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		PactLogger.Error("Unmarshal error", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	query := r.URL.Query()
	request.ProviderId = query.Get(":providerId")
	request.ConsumerId = query.Get(":consumerId")
	i, err := strconv.ParseInt(query.Get(":sha"), 10, 32)
	if err != nil {
		PactLogger.Error("Invalid pactId", err)
		controller.WriteError(w, scerr.ErrInvalidParams, err.Error())
		return
	}
	request.PactId = int32(i)
	PactLogger.Infof("PublishVerificationResults: %s, %s, %d, %t, %s\n",
		request.ProviderId, request.ConsumerId, request.PactId, request.Success,
		request.ProviderApplicationVersion)
	resp, err := ServiceAPI.PublishVerificationResults(r.Context(),
		request)
	if err != nil {
		PactLogger.Error("publish failed", err)
		controller.WriteError(w, scerr.ErrInternal, "publish failed")
		return
	}
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func (*Controller) RetrieveVerificationResults(w http.ResponseWriter, r *http.Request) {
	request := &brokerpb.RetrieveVerificationRequest{}
	query := r.URL.Query()
	request.ConsumerId = query.Get(":consumerId")
	request.ConsumerVersion = query.Get(":consumerVersion")
	PactLogger.Infof("Retrieve verification results for: %s, %s\n",
		request.ConsumerId, request.ConsumerVersion)
	resp, _ := ServiceAPI.RetrieveVerificationResults(r.Context(), request)
	respInternal := resp.Response
	resp.Response = nil
	controller.WriteResponse(w, r, respInternal, resp)
}

func getScheme(r *http.Request) string {
	if len(r.URL.Scheme) < 1 {
		return DefaultScheme
	}
	return r.URL.Scheme
}
