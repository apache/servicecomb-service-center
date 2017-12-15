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
	. "github.com/ServiceComb/service-center/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/widuu/gojson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("MicroService Api Test", func() {
	var serviceName = "integrationtestInstances"
	var serviceId = ""
	var serviceAppId = "integrationtestAppIdInstance"
	var serviceVersion = "0.0.2"
	Context("Tesing MicroService Tags API's", func() {
		BeforeEach(func() {
			schema := []string{"testSchema"}
			properties := map[string]string{"attr1": "aa"}
			servicemap := map[string]interface{}{
				"serviceName": serviceName + strconv.Itoa(rand.Int()),
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
			Expect(len(serviceId)).Should(BeNumerically("==", 32))
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

		By("Create Micro-Service Tag API", func() {
			It("Create MicroService tags", func() {
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("Create MicroService tags with invalid params", func() {
				tags := map[string]interface{}{}
				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Create MicroService tags with invalid serviceID", func() {
				tags := map[string]interface{}{
					"testkey": "testValue",
				}
				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", "XXX", 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})

		By("Get Micro-Service Tag API", func() {
			It("Get Tags for MicroService", func() {
				//Add Tag
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Get Tags
				url = strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				tagsStruct := map[string]interface{}{}
				json.Unmarshal(respbody, &tagsStruct)
				Expect(tagsStruct).To(Equal(bodyParams))
			})

			It("Get Empty Tags for MicroService", func() {
				//Get Tags
				url := strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ := scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(string(bytes.TrimSpace(respbody))).To(Equal("{}"))
			})

			It("Get Tags for Invalid MicroService", func() {
				//Get Tags
				url := strings.Replace(GETTAGS, ":serviceId", "XXXX", 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

		})
		By("Update Micro-Service Tag API", func() {
			It("Update MicroService tag with proper value", func() {
				//Add Tag
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Update Tags
				url = strings.Replace(UPDATETAG, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":key", "testkey", 1)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?value=newValue&noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Verify the Tags
				<-time.After(time.Second)
				url = strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				tagsStruct := map[string]interface{}{}
				json.Unmarshal(respbody, &tagsStruct)
				tags = map[string]interface{}{
					"testkey": "newValue",
				}

				newTags := map[string]interface{}{
					"tags": tags,
				}
				Expect(tagsStruct).To(Equal(newTags))
			})

			It("Update MicroService tag with non-exsisting tags", func() {
				//Add Tag
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Update Tags
				url = strings.Replace(UPDATETAG, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":key", "unknownkey", 1)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?value=newValue&noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			})

			It("Update MicroService tag with non-exsisting serviceID", func() {
				url := strings.Replace(UPDATETAG, ":serviceId", "XXXXX", 1)
				url = strings.Replace(url, ":key", "unknownkey", 1)
				req, _ := http.NewRequest(UPDATE, SCURL+url+"?value=newValue", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

		})
		By("Delete Micro-Service Tag API", func() {
			It("Delete MicroService tag with proper value", func() {
				//Add Tag
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Delete the tag
				url = strings.Replace(DELETETAG, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":key", "testkey", 1)
				req, _ = http.NewRequest(DELETE, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//verify Delete
				<-time.After(time.Second)
				url = strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(string(bytes.TrimSpace(respbody))).To(Equal("{}"))
			})

			It("Delete MicroService tag with non-exsisting tags", func() {
				//Add Tag
				tags := map[string]interface{}{
					"testkey": "testValue",
				}

				bodyParams := map[string]interface{}{
					"tags": tags,
				}
				url := strings.Replace(ADDTAGE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Delete the tag
				url = strings.Replace(DELETETAG, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":key", "unknowTag", 1)
				req, _ = http.NewRequest(DELETE, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				//verify Non-deleted of exsiting tag
				url = strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				tagsStruct := map[string]interface{}{}
				json.Unmarshal(respbody, &tagsStruct)
				Expect(tagsStruct).To(Equal(bodyParams))
			})

			It("Delete MicroService tag with non-exsiting service id", func() {
				//Delete the tag
				url := strings.Replace(DELETETAG, ":serviceId", "XXX", 1)
				url = strings.Replace(url, ":key", "testkey", 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
