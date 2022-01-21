package task

import (
	"context"

	"github.com/apache/servicecomb-service-center/eventbase/model"
	servicetask "github.com/apache/servicecomb-service-center/eventbase/service/task"
	carisync "github.com/go-chassis/cari/sync"
)

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
