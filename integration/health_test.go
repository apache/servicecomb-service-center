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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ServiceComb/service-center/integration"
	"github.com/widuu/gojson"
	"io/ioutil"
	"net/http"
	"testing"
)

var _ = Describe("Basic Api Test", func() {
	Context("Testing Basic Health Functions", func() {
		By("Call Health API", func() {
			It("health test", func() {
				req, _ := http.NewRequest(GET, SCURL+HEALTH, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		By("Call Version API", func() {
			It("version test", func() {
				req, _ := http.NewRequest(GET, SCURL+VERSION, nil)
				req.Header.Set("X-Domain-Name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(gojson.Json(string(respbody)).Get("apiVersion").Tostring()).To(Equal("4.0.0"))
			})
		})
	})

})

func BenchmarkHealthTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(GET, SCURL+HEALTH, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		Expect(err).To(BeNil())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
}

func BenchmarkVersionTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(GET, SCURL+VERSION, nil)
		req.Header.Set("X-Domain-Name", "default")
		resp, err := scclient.Do(req)
		Expect(err).To(BeNil())
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	}
}
