package client

import (
	"context"
	"crypto/tls"
	"testing"
)

var c = NewSyncClient("", new(tls.Config))

func TestClient_IncrementPull(t *testing.T) {
	c.IncrementPull(context.Background(), "http://127.0.0.1")
}

func TestClient_DeclareDataLength(t *testing.T) {
	c.DeclareDataLength(context.Background(), "http://127.0.0.1")
}
