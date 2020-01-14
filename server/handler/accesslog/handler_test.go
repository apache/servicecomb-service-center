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
)

func TestHandler(t *testing.T) {
	l := log.NewLogger(log.Config{})
	h := accesslog.NewAccessLogHandler(l, map[string]struct{}{})
	inv := &chain.Invocation{}
	ctx := context.Background()
	inv.Init(ctx, chain.NewChain("c", []chain.Handler{}))
	inv.WithContext(rest.CTX_MATCH_PATTERN, "/a")
	r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:80/a", nil)
	w := httptest.NewRecorder()
	inv.WithContext(rest.CTX_REQUEST, r)
	inv.WithContext(rest.CTX_RESPONSE, w)
	h.Handle(inv)
}
