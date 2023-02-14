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
	"io"
	"net/http"

	. "github.com/apache/servicecomb-service-center/integration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/widuu/gojson"
)

var _ = Describe("Syncedr Api Test", func() {
	It("health check", func() {
		req, _ := http.NewRequest(GET, SCURL+SYNCER_HEALTH, nil)
		resp, err := scclient.Do(req)
		respbody, _ := io.ReadAll(resp.Body)
		data := string(respbody)
		By("response: " + data)
		s := gojson.Json(data).Get("peers").Getindex(1).Get("status").Tostring()
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(s).To(Equal("CONNECTED"))
		defer resp.Body.Close()
	})
})
