package accesslog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/handler/accesslog"
	_ "github.com/apache/servicecomb-service-center/test"
)

func TestHandler(t *testing.T) {
	// add white list apis, ignore
	l := log.NewLogger(log.Config{})
	h := accesslog.NewAccessLogHandler(l)
	testAPI := "testAPI"
	h.AddWhiteListAPIs(testAPI)
	if !h.ShouldIgnoreAPI(testAPI) {
		t.Fatalf("Should ignore API: %s", testAPI)
	}

	// handle
	inv := &chain.Invocation{}
	ctx := context.Background()
	inv.Init(ctx, chain.NewChain("c", []chain.Handler{}))
	inv.WithContext(rest.CtxMatchPattern, "/a")
	r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:80/a", nil)
	w := httptest.NewRecorder()
	inv.WithContext(rest.CtxRequest, r)
	inv.WithContext(rest.CtxResponse, w)
	h.Handle(inv)
}
