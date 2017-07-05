package integrationtest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	. "github.com/servicecomb/service-center/tests/integrationtest"
	"github.com/widuu/gojson"
	"io/ioutil"
	"net/http"
)

var _ = Describe("Basic Api Test", func() {
	fmt.Println("Test")
	Context("Testing Basic Health Functions", func() {
		By("Call Health API", func() {
			It("health test", func() {
				req, _ := http.NewRequest(GET, SCURL+HEALTH, nil)
				req.Header.Set("X-tenant-name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(gojson.Json(string(respbody)).Get("info").Tostring()).To(Equal("Tenant mode:dedicated"))
				Expect(gojson.Json(string(respbody)).Get("status").Tostring()).To(Equal("200"))
			})
		})

		By("Call Version API", func() {
			It("version test", func() {
				req, _ := http.NewRequest(GET, SCURL+VERSION, nil)
				req.Header.Set("X-tenant-name", "default")
				resp, err := scclient.Do(req)
				Expect(err).To(BeNil())
				defer resp.Body.Close()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				respbody, _ := ioutil.ReadAll(resp.Body)
				Expect(gojson.Json(string(respbody)).Get("api_version").Tostring()).To(Equal("v3"))
			})
		})
	})

})
