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
	"net/http"
	"strings"
)

var _ = Describe("MicroService Api schema Test", func() {
	var serviceId string
	It("create microService", func() {
		schema := []string{"first_schemaId", "second_schemaId"}
		properties := map[string]string{"attr1": "aa"}
		servicemap := map[string]interface{}{
			"serviceName": "schema_test_serviceName",
			"appId":       "schema_test_appId",
			"version":     "1.0.0",
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
		respbody, _ := ioutil.ReadAll(resp.Body)
		serviceId = gojson.Json(string(respbody)).Get("serviceId").Tostring()
		Expect(err).To(BeNil())
		defer resp.Body.Close()
	})
	It("create schema", func() {
		schema := map[string]string{"schema": "first_schema"}
		url := strings.Replace(UPDATESCHEMA, ":serviceId", serviceId, -1)
		url = strings.Replace(url, ":schemaId", "first_schemaId", 1)
		body, _ := json.Marshal(schema)
		bodyBuf := bytes.NewReader(body)
		req, _ := http.NewRequest(UPDATE, SCURL+url, bodyBuf)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})

	It("create schemas", func() {
		schema := map[string]string{
			"schema":   "second_schema",
			"summary":  "second0summary",
			"schemaId": "second_schemaId",
		}
		schemas := map[string][]map[string]string{
			"schemas": []map[string]string{
				schema,
			},
		}
		url := strings.Replace(UPDATESCHEMAS, ":serviceId", serviceId, 1)
		url = strings.Replace(url, ":schemaId", "second_schemaId", 1)
		body, _ := json.Marshal(schemas)
		bodyBuf := bytes.NewReader(body)
		req, _ := http.NewRequest(POST, SCURL+url, bodyBuf)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})

	It("get schema", func() {
		url := strings.Replace(GETSCHEMABYID, ":serviceId", serviceId, 1)
		url = strings.Replace(url, ":schemaId", "second_schemaId", 1)
		req, _ := http.NewRequest(GET, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})

	It("get schemas", func() {
		url := strings.Replace(GETSCHEMAS, ":serviceId", serviceId, 1)
		req, _ := http.NewRequest(GET, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})

	It("delete schema", func() {
		url := strings.Replace(DELETESCHEMA, ":serviceId", serviceId, 1)
		url = strings.Replace(url, ":schemaId", "second_schemaId", 1)
		req, _ := http.NewRequest(DELETE, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})

	It("delete service", func() {

		url := strings.Replace(UNREGISTERMICROSERVICE, ":serviceId", serviceId, 1)
		req, _ := http.NewRequest(DELETE, SCURL+url, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, _ := scclient.Do(req)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()
	})
})
