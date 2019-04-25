package ticker

import (
	"context"
	"testing"
	"time"
)

func TestNewTaskTicker(t *testing.T) {
	ticker := NewTaskTicker(2, func(ctx context.Context) {
		t.Log("test new task ticker over")
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker.Start(ctx)
	}()
	time.Sleep(time.Second*3)
	ticker.Stop()
	cancel()
}
