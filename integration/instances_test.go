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
package integrationtest_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/apache/servicecomb-service-center/integration"
	"github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/widuu/gojson"
)

var _ = Describe("MicroService Api Test", func() {
	var serviceName = ""
	var providerID = ""
	var consumerID = ""
	var serviceAppId = "integrationtestAppIdInstance"
	var serviceVersion = "0.0.2"
	var serviceInstanceID = ""
	var alias = ""
	Context("Tesing MicroService Instances API's", func() {
		BeforeEach(func() {
			schema := []string{"testSchema"}
			properties := map[string]string{"attr1": "aa"}
			serviceName = strconv.Itoa(rand.Intn(15))
			alias = serviceAppId + ":" + serviceName
			servicemap := map[string]interface{}{
				"serviceName": serviceName,
				"alias":       alias,
				"appId":       serviceAppId,
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
			respbody, _ := io.ReadAll(resp.Body)
			providerID = gojson.Json(string(respbody)).Get("serviceId").Tostring()
			Expect(len(providerID)).Should(BeNumerically("==", LengthUUID))

			servicemap = map[string]interface{}{
				"serviceName": "C_" + serviceName,
				"appId":       serviceAppId,
				"version":     serviceVersion,
			}
			bodyParams = map[string]interface{}{
				"service": servicemap,
			}
			body, _ = json.Marshal(bodyParams)
			bodyBuf = bytes.NewReader(body)
			req, _ = http.NewRequest(POST, SCURL+REGISTERMICROSERVICE, bodyBuf)
			req.Header.Set("X-Domain-Name", "default")
			resp, err = scclient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			// Validate the service creation
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			respbody, _ = io.ReadAll(resp.Body)
			consumerID = gojson.Json(string(respbody)).Get("serviceId").Tostring()

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
			url := strings.Replace(REGISTERINSTANCE, ":serviceId", providerID, 1)
			body, _ = json.Marshal(bodyParams)
			bodyBuf = bytes.NewReader(body)
			req, _ = http.NewRequest(POST, SCURL+url, bodyBuf)
			req.Header.Set("X-Domain-Name", "default")
			resp, err = scclient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			// Validate the instance registration
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			respbody, _ = io.ReadAll(resp.Body)
			serviceInstanceID = gojson.Json(string(respbody)).Get("instanceId").Tostring()
			Expect(len(providerID)).Should(BeNumerically("==", LengthUUID))

		})

		AfterEach(func() {
			if serviceInstanceID != "" {
				url := strings.Replace(UNREGISTERINSTANCE, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			}

			if providerID != "" {
				url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", providerID, 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			}

		})

		By("Register MicroService Instance API", func() {
			It("Register MicroService Instance with invalid params", func() {
				instance := map[string]interface{}{
					"hostName":    "cse",
					"status":      "UP",
					"environment": "production",
				}

				bodyParams := map[string]interface{}{
					"instancse": instance,
				}
				url := strings.Replace(REGISTERINSTANCE, ":serviceId", providerID, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Register MicroService Instance with duplicate Params", func() {
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
					"instanceId":  serviceInstanceID,
					"endpoints":   endpoints,
					"hostName":    "cse",
					"status":      "UP",
					"environment": "production",
					"properties":  propertiesInstance,
					"healthCheck": healthcheck,
				}

				bodyParams := map[string]interface{}{
					"instance": instance,
				}
				url := strings.Replace(REGISTERINSTANCE, ":serviceId", providerID, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				// Validate the instance registration
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respbody, _ := io.ReadAll(resp.Body)

				//Validate the instance id is same as the old one
				Expect(gojson.Json(string(respbody)).Get("instanceId").Tostring()).To(Equal(serviceInstanceID))
			})

			It("Register MicroService Instance with valid params", func() {
				endpoints := []string{"cse://127.0.0.1:9985"}
				propertiesInstance := map[string]interface{}{
					"_TAGS":  "A,B",
					"attr1":  "a",
					"nodeIP": "one",
				}
				healthcheck := map[string]interface{}{
					"mode":     "push",
					"interval": 60,
					"times":    3,
				}
				instance := map[string]interface{}{
					"endpoints":   endpoints,
					"hostName":    "cse",
					"status":      "UP",
					"environment": "production",
					"properties":  propertiesInstance,
					"healthCheck": healthcheck,
				}

				bodyParams := map[string]interface{}{
					"instance": instance,
				}
				url := strings.Replace(REGISTERINSTANCE, ":serviceId", providerID, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				// Validate the instance registration
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respbody, _ := io.ReadAll(resp.Body)
				serviceInst := gojson.Json(string(respbody)).Get("instanceId").Tostring()

				//Verify the instanceID is different for two instance
				Expect(serviceInst).NotTo(Equal(serviceInstanceID))
				//Delete Instance
				url = strings.Replace(UNREGISTERINSTANCE, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInst, 1)
				req, _ = http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ = scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		By("Discover MicroService Instance API", func() {
			It("Find Micro-service Info by AppID", func() {
				req, _ := http.NewRequest(GET, SCURL+FINDINSTANCE+"?appId="+serviceAppId+"&serviceName="+serviceName+"&version="+serviceVersion, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				foundMicroServiceInstance := false
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						foundMicroServiceInstance = true
						break
					}
				}
				Expect(foundMicroServiceInstance).To(Equal(true))

				// wait dep handle
				time.Sleep(3 * time.Second)

				//Get Provider by ConsumerID
				url := strings.Replace(GETCONPRODEPENDENCY, ":consumerId", consumerID, 1)
				req, _ = http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ = scclient.Do(req)
				respbody, _ = io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &respStruct)
				Expect(providerID).To(Equal(respStruct["providers"][0]["serviceId"]))

				//Get Consumer by ProviderID
				url = strings.Replace(GETPROCONDEPENDENCY, ":providerId", providerID, 1)
				req, _ = http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ = scclient.Do(req)
				respbody, _ = io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respStruct = map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &respStruct)
				found := false
				for _, c := range respStruct["consumers"] {
					if c["serviceId"] == consumerID {
						found = true
						break
					}
				}
				Expect(true).To(Equal(found))
			})

			It("Find Micro-service Info by invalid AppID", func() {
				req, _ := http.NewRequest(GET, SCURL+FINDINSTANCE+"?appId=XXXX&serviceName="+serviceName+"&version="+serviceVersion, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Find Micro-Service Instance by ServiceID", func() {
				url := strings.Replace(GETINSTANCE, ":serviceId", providerID, 1)
				req, _ := http.NewRequest(GET, SCURL+url+"?noCache=true", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				foundMicroServiceInstance := false
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						foundMicroServiceInstance = true
						break
					}
				}
				Expect(foundMicroServiceInstance).To(Equal(true))
			})

			It("Find Micro-Service Instance by Invalid ServiceID", func() {
				url := strings.Replace(GETINSTANCE, ":serviceId", "XX", 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Find MicroServiceInstance with Service and IstanceID", func() {
				url := strings.Replace(GETINSTANCEBYINSTANCEID, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				instance := servicesStruct["instance"]
				Expect(instance["instanceId"]).To(Equal(serviceInstanceID))
			})

			It("Find Micro-Service Instance by Invalid InstanceID", func() {
				url := strings.Replace(GETINSTANCEBYINSTANCEID, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", "XX", 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Find Micro-service Info by alias", func() {
				req, _ := http.NewRequest(GET, SCURL+FINDINSTANCE+"?noCache=true&appId="+serviceAppId+"&serviceName="+alias+"&version="+serviceVersion, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				foundMicroServiceInstance := false
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						foundMicroServiceInstance = true
						break
					}
				}
				Expect(foundMicroServiceInstance).To(Equal(true))
			})

			It("Find Micro-Service Instance with rev", func() {
				req, _ := http.NewRequest(GET, SCURL+FINDINSTANCE+"?appId="+serviceAppId+"&serviceName="+serviceName+"&version="+serviceVersion, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rev := resp.Header.Get("X-Resource-Revision")
				Expect(rev).NotTo(BeEmpty())

				req, _ = http.NewRequest(GET, SCURL+FINDINSTANCE+"?appId="+serviceAppId+"&serviceName="+serviceName+"&version="+serviceVersion+"&rev="+rev, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ = scclient.Do(req)
				io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusNotModified))
				rev = resp.Header.Get("X-Resource-Revision")
				Expect(rev).NotTo(BeEmpty())
			})

			It("Batch Find Micro-service Instance", func() {
				notExistsService := map[string]interface{}{
					"service": map[string]interface{}{
						"appId":       serviceAppId,
						"serviceName": "notexisted",
						"version":     serviceVersion,
					},
				}
				provider := map[string]interface{}{
					"service": map[string]interface{}{
						"appId":       serviceAppId,
						"serviceName": serviceName,
						"version":     serviceVersion,
					},
				}
				notExistsInstance := map[string]interface{}{
					"instance": map[string]interface{}{
						"serviceId":  providerID,
						"instanceId": "notexisted",
					},
				}
				providerInstance := map[string]interface{}{
					"instance": map[string]interface{}{
						"serviceId":  providerID,
						"instanceId": serviceInstanceID,
					},
				}
				findRequest := map[string]interface{}{
					"services": []map[string]interface{}{
						provider,
						notExistsService,
					},
					"instances": []map[string]interface{}{
						providerInstance,
						notExistsInstance,
					},
				}
				body, _ := json.Marshal(findRequest)

				req, _ := http.NewRequest(POST, SCURL+INSTANCEACTION, bytes.NewReader(body))
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ := scclient.Do(req)
				io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				bodyBuf := bytes.NewReader(body)
				req, _ = http.NewRequest(POST, SCURL+INSTANCEACTION+"?type=query", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ = scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respStruct := map[string]map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &respStruct)
				servicesStruct := respStruct["services"]
				instancesStruct := respStruct["instances"]
				failed := false
				for _, services := range servicesStruct["failed"] {
					a := services["indexes"].([]interface{})[0] == 1.0
					b := services["error"].(map[string]interface{})["errorCode"] == "400012"
					if a && b {
						failed = true
						break
					}
				}
				Expect(failed).To(Equal(true))
				Expect(servicesStruct["updated"][0]["index"]).To(Equal(0.0))
				Expect(len(servicesStruct["updated"][0]["instances"].([]interface{}))).
					ToNot(Equal(0))
				failed = false
				for _, instances := range instancesStruct["failed"] {
					a := instances["indexes"].([]interface{})[0] == 1.0
					b := instances["error"].(map[string]interface{})["errorCode"] == "400017"
					if a && b {
						failed = true
						break
					}
				}
				Expect(failed).To(Equal(true))
				Expect(instancesStruct["updated"][0]["index"]).To(Equal(0.0))
				Expect(len(instancesStruct["updated"][0]["instances"].([]interface{}))).
					ToNot(Equal(0))
			})
		})

		By("Update Micro-Service Instance Information API's", func() {
			It("Update Micro-Service Instance Properties", func() {
				propertiesInstance := map[string]interface{}{
					"_TAGS":  "A,B",
					"attr1":  "NewValue",
					"nodeIP": "two",
				}
				bodyParams := map[string]interface{}{
					"properties": propertiesInstance,
				}
				url := strings.Replace(UPDATEINSTANCEMETADATA, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				time.Sleep(time.Second)
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Verify the updated properties
				url = strings.Replace(GETINSTANCE, ":serviceId", providerID, 1)
				req, _ = http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ = scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				time.Sleep(time.Second)
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						newproperties := services["properties"]
						Expect(newproperties).To(Equal(propertiesInstance))
						break
					}
				}

			})

			It("Update Micro-Service Instance Properties with invalid params", func() {
				propertiesInstance := map[string]interface{}{}
				bodyParams := map[string]interface{}{
					"properties": propertiesInstance,
				}
				url := strings.Replace(UPDATEINSTANCEMETADATA, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", "WRONGINSTANCEID", 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Update Micro-Service Instance Status", func() {
				url := strings.Replace(UPDATEINSTANCESTATUS, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(UPDATE, SCURL+url+"?value=DOWN", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				time.Sleep(time.Second)

				//Verify the Instance Status
				url = strings.Replace(GETINSTANCE, ":serviceId", providerID, 1)
				req, _ = http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ = scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				time.Sleep(time.Second)
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						newSTATUS := services["status"]
						Expect(newSTATUS).To(Equal("DOWN"))
						break
					}
				}
			})

			It("Update Micro-Service Instance Status with invalid Status", func() {
				url := strings.Replace(UPDATEINSTANCESTATUS, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(UPDATE, SCURL+url+"?value=INVALID", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				//Verify the Instance Status does not change the value
				url = strings.Replace(GETINSTANCE, ":serviceId", providerID, 1)
				req, _ = http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", consumerID)
				resp, _ = scclient.Do(req)
				respbody, _ := io.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				servicesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &servicesStruct)
				for _, services := range servicesStruct["instances"] {
					if services["instanceId"] == serviceInstanceID {
						newSTATUS := services["status"]
						Expect(newSTATUS).To(Equal("UP"))
						break
					}
				}
			})
		})

		By("Micro-Service Instance heartbeat API", func() {
			It("Send HeartBeat for micro-service instance", func() {
				url := strings.Replace(INSTANCEHEARTBEAT, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
				req, _ := http.NewRequest(UPDATE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("Send HeartBeat for wrong micro-service instance", func() {
				url := strings.Replace(INSTANCEHEARTBEAT, ":serviceId", providerID, 1)
				url = strings.Replace(url, ":instanceId", "XXX", 1)
				req, _ := http.NewRequest(UPDATE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		By("Micro-Sevrice Instance Watchern Api", func() {
			It("Call the watcher API ", func() {
				//This api gives 400 bad request for the integration test
				// as integration test is not able to make ws connection
				url := strings.Replace(INSTANCEWATCHER, ":serviceId", providerID, 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})

func BenchmarkRegisterMicroServiceInstance(b *testing.B) {
	schema := []string{"testSchema"}
	properties := map[string]string{"attr1": "aa"}
	servicemap := map[string]interface{}{
		"serviceName": "testInstance" + strconv.Itoa(rand.Int()),
		"appId":       "testApp",
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
	respbody, _ := io.ReadAll(resp.Body)
	serviceId := gojson.Json(string(respbody)).Get("serviceId").Tostring()
	Expect(len(serviceId)).Should(BeNumerically("==", LengthUUID))

	for i := 0; i < b.N; i++ {
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
		respbody, _ = io.ReadAll(resp.Body)
		serviceInstanceID := gojson.Json(string(respbody)).Get("instanceId").Tostring()
		Expect(len(serviceId)).Should(BeNumerically("==", LengthUUID))

		if serviceInstanceID != "" {
			url := strings.Replace(UNREGISTERINSTANCE, ":serviceId", serviceId, 1)
			url = strings.Replace(url, ":instanceId", serviceInstanceID, 1)
			req, _ := http.NewRequest(DELETE, SCURL+url, nil)
			req.Header.Set("X-Domain-Name", "default")
			resp, _ := scclient.Do(req)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	}
	if serviceId != "" {
		url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
		req, _ := http.NewRequest(DELETE, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
}

func BenchmarkInstanceWatch(t *testing.B) {
	scclient = insecurityConnection
	var serviceId, instanceId string

	t.Run("prepare data", func(t *testing.B) {
		// service
		serviceName := "testInstance" + strconv.Itoa(rand.Int())
		servicemap := map[string]interface{}{
			"serviceName": serviceName,
			"appId":       "testApp",
			"version":     "1.0",
		}
		bodyParams := map[string]interface{}{
			"service": servicemap,
		}
		body, _ := json.Marshal(bodyParams)
		bodyBuf := bytes.NewReader(body)
		req, _ := http.NewRequest(POST, SCURL+REGISTERMICROSERVICE, bodyBuf)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		respbody, _ := io.ReadAll(resp.Body)
		serviceId = gojson.Json(string(respbody)).Get("serviceId").Tostring()
		resp.Body.Close()

		// instance
		healthcheck := map[string]interface{}{
			"mode":     "push",
			"interval": 30000,
			"times":    20000,
		}
		instance := map[string]interface{}{
			"hostName":    "cse",
			"healthCheck": healthcheck,
		}
		bodyParams = map[string]interface{}{
			"instance": instance,
		}
		body, _ = json.Marshal(bodyParams)
		bodyBuf = bytes.NewReader(body)
		req, _ = http.NewRequest(POST, SCURL+strings.Replace(REGISTERINSTANCE, ":serviceId", serviceId, 1), bodyBuf)
		req.Header.Set("X-Domain-Name", "default")
		resp, err = scclient.Do(req)
		assert.NoError(t, err)
		respbody, _ = io.ReadAll(resp.Body)
		instanceId = gojson.Json(string(respbody)).Get("instanceId").Tostring()
		resp.Body.Close()

		req, _ = http.NewRequest(GET, SCURL+FINDINSTANCE+"?appId=testApp&serviceName="+serviceName+"&version=1.0", nil)
		req.Header.Set("X-Domain-Name", "default")
		req.Header.Set("X-ConsumerId", serviceId)
		resp, err = scclient.Do(req)
		assert.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("test 10K connection", func(t *testing.B) {

		const N, E = 2500, 2500
		var okWg sync.WaitGroup
		okWg.Add(N)

		t.Run("new 10K connection", func(t *testing.B) {
			url := strings.ReplaceAll(strings.ReplaceAll(SCURL, "http://", "ws://")+INSTANCEWATCHER, ":serviceId", serviceId)
			for i := 0; i < N; i++ {
				go func() {
					conn, _, err := websocket.DefaultDialer.Dial(url, nil)
					assert.NoError(t, err)
					for {
						_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
						_, data, err := conn.ReadMessage()
						if err != nil {
							okWg.Done()
							return
						}

						var response discovery.WatchInstanceResponse
						_ = json.Unmarshal(data, &response)
						instance := response.Instance
						timestamp, _ := strconv.ParseInt(instance.ModTimestamp, 10, 64)
						sub := time.Now().Sub(time.Unix(timestamp, 0))
						fmt.Println(instance.Properties["tag"], sub)
					}
				}()
			}
			<-time.After(10 * time.Second)
		})

		t.Run("fire 10K event", func(t *testing.B) {
			for i := 0; i < E; i++ {
				propertiesInstance := map[string]interface{}{
					"tag": strconv.Itoa(i),
				}
				bodyParams := map[string]interface{}{
					"properties": propertiesInstance,
				}
				url := strings.Replace(UPDATEINSTANCEMETADATA, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":instanceId", instanceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				resp.Body.Close()
			}
		})

		t.Run("wait", func(t *testing.B) {
			okWg.Wait()
		})
	})
}
