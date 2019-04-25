package syssig

import (
	"context"
	"syscall"
	"testing"
	"time"
)

func TestSignalsHandler(t *testing.T) {
	AddSignalsHandler(func() {},syscall.SIGHUP,syscall.SIGINT,syscall.SIGKILL,syscall.SIGTERM)
	AddSignalsHandler(func() {},syscall.Signal(999))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {Run(ctx)}()
	time.Sleep(time.Second)
	cancel()
}
