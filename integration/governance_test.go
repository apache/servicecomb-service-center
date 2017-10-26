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
package integrationtest_test

import (
	"bytes"
	"encoding/json"
	. "github.com/ServiceComb/service-center/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/widuu/gojson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

var _ = Describe("MicroService Api Test", func() {
	var serviceName = "integrationtestInstances"
	var serviceId = ""
	var serviceAppId = "integrationtestAppIdInstance"
	var serviceVersion = "0.0.2"
	var serviceInstanceID = ""
	Context("Tesing MicroService Governance API's", func() {
		BeforeEach(func() {
			schema := []string{"testSchema"}
			properties := map[string]string{"attr1": "aa"}
			servicemap := map[string]interface{}{
				"serviceName": serviceName + strconv.Itoa(rand.Int()),
				"appId":       serviceAppId + strconv.Itoa(rand.Int()),
				"version":     serviceVersion,
				"description": "examples",
				"level":       "FRONT",
				"schemas":     schema,
				"status":      "UP",
				"properties":  properties,
			}
			bodyParams := map[string]interface{}{
				"service": servicemap,
			}
			body, _ := json.Marshal(bodyParams)
			bodyBuf := bytes.NewReader(body)
			req, _ := http.NewRequest(POST, SCURL+REGISTERMICROSERVICE, bodyBuf)
			req.Header.Set("X-Domain-Name", "default")
			resp, err := scclient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			// Validate the service creation
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			respbody, _ := ioutil.ReadAll(resp.Body)
			serviceId = gojson.Json(string(respbody)).Get("serviceId").Tostring()
			Expect(len(serviceId)).Should(BeNumerically("==", 32))

			//Register MicroService Instance
			endpoints := []string{"cse://127.0.0.1:9984"}
			propertiesInstance := map[string]interface{}{
				"_TAGS":  "A,B",
				"attr1":  "a",
				"nodeIP": "one",
			}
			healthcheck := map[string]interface{}{
				"mode":     "push",
				"interval": 30,
				"times":    2,
			}
			instance := map[string]interface{}{
				"endpoints":   endpoints,
				"hostName":    "cse",
				"status":      "UP",
				"environment": "production",
				"properties":  propertiesInstance,
				"healthCheck": healthcheck,
			}

			bodyParams = map[string]interface{}{
				"instance": instance,
			}
			url := strings.Replace(REGISTERINSTANCE, ":serviceId", serviceId, 1)
			body, _ = json.Marshal(bodyParams)
			bodyBuf = bytes.NewReader(body)
			req, _ = http.NewRequest(POST, SCURL+url, bodyBuf)
			req.Header.Set("X-Domain-Name", "default")
			resp, err = scclient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			// Validate the instance registration
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			respbody, _ = ioutil.ReadAll(resp.Body)
			serviceInstanceID = gojson.Json(string(respbody)).Get("instanceId").Tostring()
			Expect(len(serviceId)).Should(BeNumerically("==", 32))

		})

		AfterEach(func() {
			if serviceInstanceID != "" {
				url := strings.Replace(UNREGISTERINSTANCE, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			}

			if serviceId != "" {
				url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			}

		})

		By("Get Service Information using governance API", func() {
			It("Get ServiceInfo by Governance API", func() {
				url := strings.Replace(GETGOVERNANCESERVICEDETAILS, ":serviceId", serviceId, 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string]map[string]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				microservice := servicesStruct["service"]
				microserviceID := microservice["microSerivce"]
				Expect(microserviceID["serviceId"]).To(Equal(serviceId))
			})

			It("Get ServiceInfo by Governance API with non-exsistence serviceID", func() {
				url := strings.Replace(GETGOVERNANCESERVICEDETAILS, ":serviceId", "XXXXX", 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		By("GET Relation Graph for all microservice", func() {
			It("Get Relation Graph for all ", func() {
				req, _ := http.NewRequest(GET, SCURL+GETRELATIONGRAPH, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				relationStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &relationStruct)
				foundMicroService := false
				for _, services := range relationStruct["nodes"] {
					if services["id"] == serviceId {
						foundMicroService = true
						break
					}
				}
				Expect(foundMicroService).To(Equal(true))
			})
		})

		By("Get All Services information by Governance API", func() {
			It("Get All Service Metadata", func() {
				req, _ := http.NewRequest(GET, SCURL+GETALLSERVICEGOVERNANCEINFO, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				relationStruct := map[string][]map[string]map[string]interface{}{}

				json.Unmarshal(respbody, &relationStruct)
				foundMicroService := false
				for _, microservices := range relationStruct["allServicesDetail"] {
					service := microservices["microSerivce"]
					if service["serviceId"] == serviceId {
						foundMicroService = true
						break
					}
				}
				Expect(foundMicroService).To(Equal(true))

			})
		})
	})

})

func BenchmarkGovernance(b *testing.B) {
	schema := []string{"testSchema"}
	properties := map[string]string{"attr1": "aa"}
	servicemap := map[string]interface{}{
		"serviceName": "testGov",
		"appId":       "testGovID",
		"version":     "1.0",
		"description": "examples",
		"level":       "FRONT",
		"schemas":     schema,
		"status":      "UP",
		"properties":  properties,
	}
	bodyParams := map[string]interface{}{
		"service": servicemap,
	}
	body, _ := json.Marshal(bodyParams)
	bodyBuf := bytes.NewReader(body)
	req, _ := http.NewRequest(POST, SCURL+REGISTERMICROSERVICE, bodyBuf)
	req.Header.Set("X-Domain-Name", "default")
	resp, err := scclient.Do(req)
	Expect(err).To(BeNil())
	defer resp.Body.Close()

	// Validate the service creation
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	respbody, _ := ioutil.ReadAll(resp.Body)
	serviceId := gojson.Json(string(respbody)).Get("serviceId").Tostring()
	Expect(len(serviceId)).Should(BeNumerically("==", 32))

	for i := 0; i < b.N; i++ {
		url := strings.Replace(GETGOVERNANCESERVICEDETAILS, ":serviceId", serviceId, 1)
		req, _ := http.NewRequest(GET, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		Expect(err).To(BeNil())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
	if serviceId != "" {
		url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
		req, _ := http.NewRequest(DELETE, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
}
