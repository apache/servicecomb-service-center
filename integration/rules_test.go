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
	. "github.com/apache/incubator-servicecomb-service-center/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/widuu/gojson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

var _ = Describe("MicroService Api Test", func() {
	var serviceName = "integrationtestInstances"
	var serviceId = ""
	var serviceAppId = "integrationtestAppIdInstance"
	var serviceVersion = "0.0.2"
	var ruleID = ""
	Context("Tesing MicroService Rules API's", func() {
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

		By("Create Micro-Service Rules API", func() {
			It("Create MicroService Rules", func() {
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("Create MicroService Rules with empty rules", func() {
				rules := map[string]interface{}{}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Create MicroService Rules with wrong Rule Type", func() {
				rules := map[string]interface{}{
					"ruleType":    "UKNOWN",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Create MicroService Rules with worng service ID", func() {
				rules := map[string]interface{}{}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", "XXXXXX", 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Create MicroService Rules with duplicate rules", func() {
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Duplicate Request
				bodyBuf = bytes.NewReader(body)
				req, _ = http.NewRequest(POST, SCURL+url+"?noCache=1", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(strings.TrimSpace(string(respbody))).To(Equal("{}"))
			})
		})

		By("Get Micro-Service Rules API", func() {
			It("Get Rules for MicroService", func() {
				//Add Rules
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Get Rules
				url = strings.Replace(GETRULES, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rulesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &rulesStruct)
				for _, rule := range rulesStruct["rules"] {
					Expect(rule["ruleType"]).To(Equal("WHITE"))
					Expect(rule["attribute"]).To(Equal("tag_a"))
					Expect(rule["pattern"]).To(Equal("a+"))
				}
			})

			It("Get Empty Rules for MicroService", func() {
				//Get Rules
				url := strings.Replace(GETRULES, ":serviceId", serviceId, 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ := scclient.Do(req)
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(strings.TrimSpace(string(respbody))).To(Equal("{}"))
			})

			It("Get Rules for Invalid MicroService", func() {
				//Get Rules
				url := strings.Replace(GETRULES, ":serviceId", "XXXX", 1)
				req, _ := http.NewRequest(GET, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, _ := scclient.Do(req)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

		})
		By("Update Micro-Service Rules API", func() {
			It("Update MicroService rules with proper value", func() {
				//Add Rules
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				//Get the rule ID
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rulesCreateRespStruct := map[string][]string{}
				json.Unmarshal(respbody, &rulesCreateRespStruct)
				for _, rule := range rulesCreateRespStruct["RuleIds"] {
					ruleID = rule
					break
				}

				//Update Rules
				updateParams := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_b",
					"pattern":     "a+",
					"description": "test ",
				}
				url = strings.Replace(UPDATERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
				body, _ = json.Marshal(updateParams)
				bodyBuf = bytes.NewReader(body)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?noCache=1", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				respbody, _ = ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//Get Rules
				url = strings.Replace(GETRULES, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ = ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rulesStruct := map[string][]map[string]interface{}{}
				json.Unmarshal(respbody, &rulesStruct)
				for _, rule := range rulesStruct["rules"] {
					Expect(rule["ruleType"]).To(Equal("WHITE"))
					Expect(rule["attribute"]).To(Equal("tag_b"))
					Expect(rule["pattern"]).To(Equal("a+"))
				}
			})

			It("Update MicroService tag with invalid Rules", func() {
				//Add Rules
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				//Get the rule ID
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rulesCreateRespStruct := map[string][]string{}
				json.Unmarshal(respbody, &rulesCreateRespStruct)
				for _, rule := range rulesCreateRespStruct["RuleIds"] {
					ruleID = rule
					break
				}

				//Update Rules with invalid RuleType
				updateParams := map[string]interface{}{
					"ruleType":    "UNKNOWN",
					"attribute":   "tag_b",
					"pattern":     "a+",
					"description": "test ",
				}
				url = strings.Replace(UPDATERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
				body, _ = json.Marshal(updateParams)
				bodyBuf = bytes.NewReader(body)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?noCache=1", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				//Update Rules with invalid pattern
				updateParams = map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_b",
					"pattern":     "",
					"description": "test ",
				}
				url = strings.Replace(UPDATERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
				body, _ = json.Marshal(updateParams)
				bodyBuf = bytes.NewReader(body)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?noCache=1", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				//Update Rules with invalid different ruleType
				updateParams = map[string]interface{}{
					"ruleType":    "BLACK",
					"attribute":   "tag_b",
					"pattern":     "a+",
					"description": "test ",
				}
				url = strings.Replace(UPDATERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
				body, _ = json.Marshal(updateParams)
				bodyBuf = bytes.NewReader(body)
				req, _ = http.NewRequest(UPDATE, SCURL+url+"?noCache=1", bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			})

			It("Update MicroService Rule with non-exsisting RuleID", func() {
				updateParams := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_b",
					"pattern":     "a+",
					"description": "test ",
				}
				url := strings.Replace(UPDATERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", "XXXX", 1)
				body, _ := json.Marshal(updateParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

		})
		By("Delete Micro-Service Rules API", func() {
			It("Delete MicroService Rules with proper value", func() {
				//Add Rules
				rules := map[string]interface{}{
					"ruleType":    "WHITE",
					"attribute":   "tag_a",
					"pattern":     "a+",
					"description": "test ",
				}
				rulesArray := []interface{}{rules}
				bodyParams := map[string][]interface{}{
					"rules": rulesArray,
				}
				url := strings.Replace(ADDRULE, ":serviceId", serviceId, 1)
				body, _ := json.Marshal(bodyParams)
				bodyBuf := bytes.NewReader(body)
				req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()

				//Get the rule ID
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				rulesCreateRespStruct := map[string][]string{}
				json.Unmarshal(respbody, &rulesCreateRespStruct)
				for _, rule := range rulesCreateRespStruct["RuleIds"] {
					ruleID = rule
					break
				}

				//Delete the Rules
				url = strings.Replace(DELETERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
				req, _ = http.NewRequest(DELETE, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err = scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				//verify Delete
				url = strings.Replace(GETTAGS, ":serviceId", serviceId, 1)
				req, _ = http.NewRequest(GET, SCURL+url+"?noCache=1", nil)
				req.Header.Set("X-Domain-Name", "default")
				req.Header.Set("X-ConsumerId", serviceId)
				resp, _ = scclient.Do(req)
				respbody, _ = ioutil.ReadAll(resp.Body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(strings.TrimSpace(string(respbody))).To(Equal("{}"))
			})

			It("Delete MicroService rules with non-exsisting ruleID", func() {
				url := strings.Replace(DELETERULES, ":serviceId", serviceId, 1)
				url = strings.Replace(url, ":rule_id", "XX", 1)
				req, _ := http.NewRequest(DELETE, SCURL+url, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("Delete MicroService rules with non-exsiting service id", func() {
				url := strings.Replace(DELETERULES, ":serviceId", "XX", 1)
				url = strings.Replace(url, ":rule_id", ruleID, 1)
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
