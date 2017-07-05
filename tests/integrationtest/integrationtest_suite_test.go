package integrationtest_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Integration Test for SC", []Reporter{junitReporter})
}

var scclient *http.Client

var insecurityConnection = &http.Client{}

var SCURL = "http://127.0.0.1:30100"

var _ = BeforeSuite(func() {
	scclient = insecurityConnection
})
