package sync

import (
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
)

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator"
	"github.com/apache/servicecomb-service-center/syncer/service/task"

	"github.com/go-chassis/foundation/gopool"
)

func Init() {
	if !config.GetConfig().Sync.EnableOnStart {
		log.Info("sync not enabled")
		return
	}

	gopool.Go(func(ctx context.Context) {
		Work()
	})
}

func Work() {
	work()
}

func work() {
	err := replicator.Work()
	if err != nil {
		log.Warn(err.Error())
		return
	}

	event.Work()

	task.Work()
}
