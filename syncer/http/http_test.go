package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHttpServer(t *testing.T) {
	addr := "127.0.0.1:9099"
	svr, err := NewServer(
		WithAddr(addr),
		WithTLSConfig(nil),
	)
	assert.NoError(t, err, "create server succeed")

	t.Run("server start test", func(t *testing.T) {
		err := startServer(context.Background(), svr)
		assert.NoError(t, err, "server start succeed")
	})

	t.Run("server stop test", func(t *testing.T) {
		svr.Stop()
	})
}

func startServer(ctx context.Context, svr *Server) (err error) {
	svr.Start(ctx)
	select {
	case <-svr.Ready():
	case <-svr.Stopped():
		err = errors.New("start http server failed")
	case <-time.After(time.Second * 3):
		err = errors.New("start http server timeout")
	}
	return
}
