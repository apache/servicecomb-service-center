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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"

	. "github.com/apache/servicecomb-service-center/integration"
	"github.com/widuu/gojson"
)

var serviceName = ""

var _ = Describe("MicroService Api Test", func() {
	var serviceId = ""
	var serviceAppId = "integrationtestAppId"
	var serviceVersion = "0.0.1"
	Context("Testing MicroServices Functions", func() {
		By("Test Register Micro-Service API", func() {
			It("Register MicroService", func() {
				schema := []string{"testSchema"}
				properties := map[string]string{"attr1": "aa"}
				serviceName = strconv.Itoa(rand.Intn(15))
				servicemap := map[string]interface{}{
					"serviceName": serviceName,
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
				respbody, _ := ioutil.ReadAll(resp.Body)
				serviceId = gojson.Json(string(respbody)).Get("serviceId").Tostring()
				Expect(len(serviceId)).Should(BeNumerically("==", LengthUUID))

				// UNRegister Service
				url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ = scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

			})
		})

		Context("Testing MicroService API's", func() {
			BeforeEach(func() {
				schema := []string{"testSchema"}
				properties := map[string]string{"attr1": "aa"}
				serviceName = strconv.Itoa(rand.Intn(15))
				servicemap := map[string]interface{}{
					"serviceName": serviceName,
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
				respbody, _ := ioutil.ReadAll(resp.Body)
				serviceId = gojson.Json(string(respbody)).Get("serviceId").Tostring()
				Expect(len(serviceId)).Should(BeNumerically("==", LengthUUID))
			})

			AfterEach(func() {
				if serviceId != "" {
					url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
					req, _ := http.NewRequest(DELETE, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				}

			})
			By("Test Micro-Service Unregister API", func() {
				It("test valid scenario", func() {
					url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
					req, _ := http.NewRequest(DELETE, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					serviceId = ""
				})

				It("test invalid id", func() {
					url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", "56565", 1)
					req, _ := http.NewRequest(DELETE, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})
			})

			By("Test if Service Exsist", func() {
				It("test valid scenario", func() {
					getServiceName(serviceId)
					req, _ := http.NewRequest(GET, SCURL+CHECKEXISTENCE+"?type=microservice&appId="+serviceAppId+"&serviceName="+serviceName+"&version="+serviceVersion, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					respbody, _ := ioutil.ReadAll(resp.Body)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(gojson.Json(string(respbody)).Get("serviceId").Tostring()).To(Equal(serviceId))
				})

				It("test invalid api params", func() {
					req, _ := http.NewRequest(GET, SCURL+CHECKEXISTENCE+"?type=microservice&appId=&serviceName="+serviceName+"&version="+serviceVersion, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})
			})

			By("Test Get Service by ServiceID", func() {
				It("test valid scenario", func() {
					getServiceName(serviceId)
					url := strings.Replace(GETSERVICEBYID, ":serviceId", serviceId, 1)
					req, _ := http.NewRequest(GET, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					respbody, _ := ioutil.ReadAll(resp.Body)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					serviceRetrived := (gojson.Json(string(respbody)).Get("service")).Getdata()
					Expect(serviceRetrived["serviceName"]).To(Equal(serviceName))
				})

				It("test invalid api params", func() {
					url := strings.Replace(GETSERVICEBYID, ":serviceId", "abcd", 1)
					req, _ := http.NewRequest(GET, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})
			})

			By("Test Get All Services in SC", func() {
				It("test valid scenario", func() {
					getServiceName(serviceId)
					req, _ := http.NewRequest(GET, SCURL+GETALLSERVICE, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					respbody, _ := ioutil.ReadAll(resp.Body)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					servicesStruct := map[string][]map[string]interface{}{}
					json.Unmarshal(respbody, &servicesStruct)
					foundMicroService := false
					for _, services := range servicesStruct["services"] {
						if services["serviceName"] == serviceName {
							foundMicroService = true
							break
						}
					}
					Expect(foundMicroService).To(Equal(true))
				})

				It("test get all service with wrong domain name", func() {
					req, _ := http.NewRequest(GET, SCURL+GETALLSERVICE, nil)
					req.Header.Set("X-Domain-Name", "default1")
					resp, _ := scclient.Do(req)
					respbody, _ := ioutil.ReadAll(resp.Body)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(strings.TrimSpace(string(respbody))).To(Equal("{}"))
				})
			})

			By("Test Update MicroService Properties", func() {
				It("test valid scenario", func() {
					url := strings.Replace(UPDATEMICROSERVICE, ":serviceId", serviceId, 1)
					properties := map[string]string{"attr2": "bb"}
					bodyParams := map[string]interface{}{
						"properties": properties,
					}
					body, _ := json.Marshal(bodyParams)
					bodyBuf := bytes.NewReader(body)
					req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
				})

				It("test invalid scenario with wrong serviceID", func() {
					url := strings.Replace(UPDATEMICROSERVICE, ":serviceId", "WrongID", 1)
					properties := map[string]string{"attr2": "bb"}
					bodyParams := map[string]interface{}{
						"properties": properties,
					}
					body, _ := json.Marshal(bodyParams)
					bodyBuf := bytes.NewReader(body)
					req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})

				It("test invalid scenario with wrong propertyType", func() {
					url := strings.Replace(UPDATEMICROSERVICE, ":serviceId", "WrongID", 1)
					properties := map[string]string{"attr2": "bb"}
					bodyParams := map[string]interface{}{
						"wrongtype": properties,
					}
					body, _ := json.Marshal(bodyParams)
					bodyBuf := bytes.NewReader(body)
					req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				})
			})

			By("Test Dependency API for Provider and Consumer", func() {
				It("Invalid scenario for GET Providers and Consumers", func() {
					//Get Provider by ConsumerID
					url := strings.Replace(GETCONPRODEPENDENCY, ":consumerId", "wrongID", 1)
					req, _ := http.NewRequest(GET, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ := scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

					//Get Consumer by ProviderID
					url = strings.Replace(GETPROCONDEPENDENCY, ":providerId", "wrongID", 1)
					req, _ = http.NewRequest(GET, SCURL+url, nil)
					req.Header.Set("X-Domain-Name", "default")
					resp, _ = scclient.Do(req)
					Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				})
			})

		})
	})

})

func BenchmarkRegisterMicroServiceAndDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		schema := []string{"testSchema"}
		properties := map[string]string{"attr1": "aa"}
		servicemap := map[string]interface{}{
			"serviceName": "test" + strconv.Itoa(i),
			"appId":       "testApp" + strconv.Itoa(i),
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
		Expect(len(serviceId)).Should(BeNumerically("==", LengthUUID))
		if serviceId != "" {
			url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
			req, _ := http.NewRequest(DELETE, SCURL+url, nil)
			req.Header.Set("X-Domain-Name", "default")
			resp, _ := scclient.Do(req)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	}
}

func BenchmarkRegisterMicroService(b *testing.B) {
	for i := 0; i < b.N; i++ {
		schema := []string{"testSchema"}
		properties := map[string]string{"attr1": "aa"}
		servicemap := map[string]interface{}{
			"serviceName": "testnd" + strconv.Itoa(rand.Int()) + strconv.Itoa(i),
			"appId":       "nd" + strconv.Itoa(rand.Int()) + strconv.Itoa(i),
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
		_, err := scclient.Do(req)
		Expect(err).To(BeNil())
	}
}

func getServiceName(serviceId string) {
	url := strings.Replace(GETSERVICEBYID, ":serviceId", serviceId, 1)
	req, _ := http.NewRequest(GET, SCURL+url, nil)
	req.Header.Set("X-Domain-Name", "default")
	resp, _ := scclient.Do(req)
	respbody, _ := ioutil.ReadAll(resp.Body)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	serviceRetrived := (gojson.Json(string(respbody)).Get("service")).Getdata()
	serviceName = serviceRetrived["serviceName"].(string)
}
