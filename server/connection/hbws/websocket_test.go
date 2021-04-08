package hbws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/server/connection/hbws"
	"github.com/apache/servicecomb-service-center/server/core"
)

var closeCh = make(chan struct{})

func init() {
	testing.Init()
	core.Initialize()
}

type watcherConn struct {
	clientConn *websocket.Conn
	serverConn *websocket.Conn
}

func (h *watcherConn) Test() {
	s := httptest.NewServer(h)
	h.clientConn, _, _ = websocket.DefaultDialer.Dial(
		strings.Replace(s.URL, "http://", "ws://", 1), nil)
}

func (h *watcherConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	h.serverConn, _ = upgrader.Upgrade(w, r, nil)
	for {
		//h.ServerConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second))
		//h.ServerConn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
		_, _, err := h.serverConn.ReadMessage()
		if err != nil {
			return
		}
		<-closeCh
		h.serverConn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
		h.serverConn.Close()
		return
	}
}

func NewTest() *watcherConn {
	ts := &watcherConn{}
	ts.Test()
	return ts
}

func TestHeartbeat(t *testing.T) {
	mock := NewTest()
	go hbws.Heartbeat(context.Background(), mock.serverConn, "", "")
	err := mock.serverConn.WriteMessage(websocket.TextMessage, []byte("hello"))
	assert.Nil(t, err)
	_, p, err := mock.clientConn.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, "hello", string(p))
}
