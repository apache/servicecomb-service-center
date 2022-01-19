package task

import (
	"context"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	servicetask "github.com/apache/servicecomb-service-center/eventbase/service/task"
	"github.com/apache/servicecomb-service-center/pkg/log"
	serverconfig "github.com/apache/servicecomb-service-center/server/config"

	carisync "github.com/go-chassis/cari/sync"
)

func initDatabase() {
	kind := serverconfig.GetString("registry.kind", "",
		serverconfig.WithStandby("registry_plugin"))

	if err := datasource.Init(kind); err != nil {
		log.Fatal("init datasource failed", err)
	}
}

func ListTask(ctx context.Context) ([]*carisync.Task, error) {
	return servicetask.List(ctx, &model.ListTaskRequest{})
}

type syncTasks []*carisync.Task

func (s syncTasks) Len() int {
	return len(s)
}

func (s syncTasks) Less(i, j int) bool {
	return s[i].Timestamp <= s[j].Timestamp
}

func (s syncTasks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
