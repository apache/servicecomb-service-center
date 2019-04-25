package ticker

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"sync"
	"time"
)

type TaskTicker struct {
	interval time.Duration
	handler  func(ctx context.Context)
	once     sync.Once
	ticker   *time.Ticker
}

func NewTaskTicker(interval int, handler func(ctx context.Context)) *TaskTicker {
	return &TaskTicker{
		interval: time.Second * time.Duration(interval),
		handler:  handler,
	}
}

func (t *TaskTicker) Start(ctx context.Context) {
	t.once.Do(func() {
		t.handler(ctx)
		t.ticker = time.NewTicker(t.interval)
		for {
			select {
			case <-t.ticker.C:
				t.handler(ctx)
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	})
}

func (t *TaskTicker) Stop() {
	if t.ticker == nil{
		return
	}
	t.ticker.Stop()
	log.Info("ticker stop")
	t.ticker = nil
}
