package task

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/service/event"

	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	receiver := make(chan struct{}, 1)
	fs := &mockSender{
		events:  make(map[string]*event.Event),
		receive: receiver,
	}
	ctx := context.TODO()
	m := NewManager(
		ManagerOperator(&mockOperator{
			tasks: map[string]*sync.Task{
				"xxx1": {
					ID:           "xxx1",
					ResourceType: "demo",
					Action:       "create",
					Timestamp:    0,
					Status:       "pending",
				},
			},
		}),
		ManagerInternal(defaultInternal),
		EventSender(fs))

	m.LoadAndHandleTask(ctx)
	m.UpdateResultTask(ctx)
	<-receiver
	assert.Equal(t, 1, len(fs.events))
}

type mockOperator struct {
	tasks map[string]*sync.Task
}

func (f *mockOperator) ListTasks(_ context.Context) ([]*sync.Task, error) {
	result := make([]*sync.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		result = append(result, task)
	}
	return result, nil
}

func (f *mockOperator) DeleteTask(_ context.Context, t *sync.Task) error {
	delete(f.tasks, t.ID)
	return nil
}

type mockSender struct {
	events  map[string]*event.Event
	receive chan struct{}
}

func (f *mockSender) Send(et *event.Event) {
	f.events[et.Id] = et
	f.receive <- struct{}{}
}
