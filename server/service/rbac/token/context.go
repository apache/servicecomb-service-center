package token

import (
	"context"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/util"
)

const CtxRequestToken util.CtxKey = "_request_token"

func WithRequest(req *http.Request, token string) *http.Request {
	return util.SetRequestContext(req, CtxRequestToken, token)
}

func FromContext(ctx context.Context) string {
	token, ok := ctx.Value(CtxRequestToken).(string)
	if !ok {
		return ""
	}
	return token
}
