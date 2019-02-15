package sc

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func (c *LBClient) WebsocketDial(ctx context.Context, api string, headers http.Header) (conn *websocket.Conn, err error) {
	dialer := &websocket.Dialer{TLSClientConfig: c.TLS}
	var errs []string
	for i := 0; i < c.Retries; i++ {
		var addr *url.URL
		addr, err = url.Parse(c.Next())
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s]: %s", addr, err.Error()))
			continue
		}
		if addr.Scheme == "https" {
			addr.Scheme = "wss"
		} else {
			addr.Scheme = "ws"
		}
		conn, _, err = dialer.Dial(addr.String() + api, headers)
		if err == nil {
			break
		}
		errs = append(errs, fmt.Sprintf("[%s]: %s", addr, err.Error()))
	}
	if err != nil {
		err = errors.New(util.StringJoin(errs, ", "))
	}
	return
}
