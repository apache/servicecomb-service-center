package sync

import (
	"fmt"

	// glint
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator"
	"github.com/apache/servicecomb-service-center/syncer/service/task"
)

func Init() {
	if !config.GetConfig().Sync.EnableOnStart {
		log.Info("sync not enabled")
		return
	}

	Work()
}

func Work() {
	work()
}

func work() {
	err := replicator.Work()
	if err != nil {
		log.Warn(fmt.Sprintf("replicate work init failed, %s", err.Error()))
		return
	}

	event.Work()

	task.Work()
}
