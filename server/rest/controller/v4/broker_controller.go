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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/rest/controller"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
)

const DEFAULT_SCHEME = "http"

func getScheme(r *http.Request) string {
	if len(r.URL.Scheme) < 1 {
		return DEFAULT_SCHEME
	}
	return r.URL.Scheme
}

func init() {

}

type BrokerService struct {
	//
}

func (this *BrokerService) URLPatterns() []rest.Route {
	return []rest.Route{
		// for handling broker requests
		{rest.HTTP_METHOD_GET,
			"/broker",
			this.GetHome},
		{rest.HTTP_METHOD_PUT,
			"/broker/pacts/provider/:providerId/consumer/:consumerId/version/:number",
			this.PublishPact},
		{rest.HTTP_METHOD_GET,
			"/broker/pacts/provider/:providerId/latest",
			this.GetAllProviderPacts},
		{rest.HTTP_METHOD_GET,
			"/broker/pacts/provider/:providerId/consumer/:consumerId/version/:number",
			this.GetPactsOfProvider},
		{rest.HTTP_METHOD_DELETE,
			"/broker/pacts/delete",
			this.DeletePacts},
		{rest.HTTP_METHOD_POST,
			"/broker/pacts/provider/:providerId/consumer/:consumerId/pact-version/:sha/verification-results",
			this.PublishVerificationResults},
		{rest.HTTP_METHOD_GET,
			"/broker/verification-results/consumer/:consumerId/version/:consumerVersion/latest",
			this.RetrieveVerificationResults},
	}
}

func (s *BrokerService) GetHome(w http.ResponseWriter, r *http.Request) {
	request := &pb.BaseBrokerRequest{
		HostAddress: r.Host,
		Scheme:      getScheme(r),
	}
	resp, _ := core.BrokerServiceAPI.GetBrokerHome(r.Context(), request)
	if resp.GetResponse().GetCode() != pb.Response_SUCCESS {
		controller.WriteJson(http.StatusBadRequest, []byte(resp.Response.GetMessage()), w)
	} else {
		controller.WriteJsonObject(http.StatusOK, resp, w)
	}

}

func (this *BrokerService) GetAllProviderPacts(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetAllProviderPactsRequest{
		ProviderId: r.URL.Query().Get(":providerId"),
		BaseUrl: &pb.BaseBrokerRequest{
			HostAddress: r.Host,
			Scheme:      getScheme(r),
		},
	}
	//href := GenerateBrokerAPIPath(r, "/broker/pacts/provider/"+request.ProviderId+"/consumer/", nil)

	//serviceUtil.PactLogger.Infof("########################################## GetAllProviderPacts: %s\n", href)

	resp, err := core.BrokerServiceAPI.GetAllProviderPacts(r.Context(), request /*, href*/)
	linksObj, err := json.Marshal(resp)
	if err != nil {
		return
	}
	serviceUtil.PactLogger.Infof("Pact info: %s\n", string(linksObj))
	controller.WriteJsonResponse(resp.GetResponse(), resp, err, w)
}

func (this *BrokerService) GetPactsOfProvider(w http.ResponseWriter, r *http.Request) {
	request := &pb.GetProviderConsumerVersionPactRequest{
		ProviderId: r.URL.Query().Get(":providerId"),
		ConsumerId: r.URL.Query().Get(":consumerId"),
		Version:    r.URL.Query().Get(":number"),
		BaseUrl: &pb.BaseBrokerRequest{
			HostAddress: r.Host,
			Scheme:      getScheme(r),
		},
	}

	resp, _ := core.BrokerServiceAPI.GetPactsOfProvider(r.Context(), request)

	if resp.GetResponse().GetCode() != pb.Response_SUCCESS {
		controller.WriteJson(http.StatusBadRequest, []byte(resp.Response.GetMessage()), w)
	} else {
		controller.WriteJsonObject(http.StatusOK, resp.Pact, w)
	}

}

func (this *BrokerService) PublishPact(w http.ResponseWriter, r *http.Request) {
	message, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serviceUtil.PactLogger.Error("body err\n", err)
		controller.WriteText(http.StatusBadRequest, err.Error(), w)
		return
	}
	request := &pb.PublishPactRequest{
		ProviderId: r.URL.Query().Get(":providerId"),
		ConsumerId: r.URL.Query().Get(":consumerId"),
		Version:    r.URL.Query().Get(":number"),
		Pact:       message,
	}
	serviceUtil.PactLogger.Infof("PublishPact: providerId = %s, consumerId = %s, version = %s\n",
		request.ProviderId, request.ConsumerId, request.Version)
	resp, err := core.BrokerServiceAPI.PublishPact(r.Context(), request)
	controller.WriteTextResponse(resp.GetResponse(), err, "pact publish success", w)
}

func (this *BrokerService) DeletePacts(w http.ResponseWriter, r *http.Request) {
	_, err := core.BrokerServiceAPI.DeletePacts(r.Context(), &pb.BaseBrokerRequest{
		HostAddress: r.Host,
		Scheme:      getScheme(r),
	})
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "Pact delete request failed\n")
		controller.WriteText(http.StatusBadRequest, "", w)
	}
}

func (this *BrokerService) PublishVerificationResults(w http.ResponseWriter, r *http.Request) {
	serviceUtil.PactLogger.Infof("########################################## Pact broker received verification results\n")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serviceUtil.PactLogger.Error("body err", err)
		controller.WriteText(http.StatusBadRequest, err.Error(), w)
		return
	}
	request := &pb.PublishVerificationRequest{}
	err = json.Unmarshal(requestBody, request)
	if err != nil {
		serviceUtil.PactLogger.Error("Invalid json", err)
		controller.WriteText(http.StatusInternalServerError, "Unmarshal error", w)
		return
	}
	request.ProviderId = r.URL.Query().Get(":providerId")
	request.ConsumerId = r.URL.Query().Get(":consumerId")
	i, err := strconv.ParseInt(r.URL.Query().Get(":sha"), 10, 32)
	if err != nil {
		serviceUtil.PactLogger.Error("Invalid pactId", err)
		controller.WriteText(http.StatusInternalServerError, "input read error", w)
		return
	}
	request.PactId = int32(i)
	serviceUtil.PactLogger.Infof("PublishVerificationResults: %s, %s, %d, %t, %s\n",
		request.ProviderId, request.ConsumerId, request.PactId, request.Success,
		request.ProviderApplicationVersion)
	resp, err := core.BrokerServiceAPI.PublishVerificationResults(r.Context(),
		request)
	controller.WriteTextResponse(resp.GetResponse(), err, "verification results published successfully", w)
}

func (this *BrokerService) RetrieveVerificationResults(w http.ResponseWriter, r *http.Request) {
	serviceUtil.PactLogger.Infof("########################################## Pact broker received retrieve verification results request\n")
	request := &pb.RetrieveVerificationRequest{}
	request.ConsumerId = r.URL.Query().Get(":consumerId")
	request.ConsumerVersion = r.URL.Query().Get(":consumerVersion")
	serviceUtil.PactLogger.Infof("Retrieve verification results for: %s, %s\n",
		request.ConsumerId, request.ConsumerVersion)
	resp, err := core.BrokerServiceAPI.RetrieveVerificationResults(r.Context(), request)
	controller.WriteJsonResponse(resp.GetResponse(), resp, err, w)
	//WriteTextResponse(resp.GetResponse(), err, "verification results retrieved successfully", w)
}
