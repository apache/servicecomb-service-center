package event

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"

	"github.com/go-chassis/foundation/gopool"
)

const (
	defaultInternal = 500 * time.Millisecond
)

var m Manager

type Event struct {
	*v1sync.Event

	Result chan<- *Result
}

type Result struct {
	ID    string
	Data  *v1sync.Result
	Error error
}

func Work() {
	m = NewManager()
	m.HandleEvent()
	m.HandleResult()
}

func GetManager() Manager {
	return m
}

type ManagerOption func(*managerOptions)

type managerOptions struct {
	internal   time.Duration
	replicator replicator.Replicator
}

func ManagerInternal(i time.Duration) ManagerOption {
	return func(options *managerOptions) {
		options.internal = i
	}
}

func toManagerOptions(os ...ManagerOption) *managerOptions {
	mo := new(managerOptions)
	mo.internal = defaultInternal
	mo.replicator = replicator.Manager()
	for _, o := range os {
		o(mo)
	}

	return mo
}

func Replicator(r replicator.Replicator) ManagerOption {
	return func(options *managerOptions) {
		options.replicator = r
	}
}

func NewManager(os ...ManagerOption) Manager {
	mo := toManagerOptions(os...)
	em := &eventManager{
		events:     make(chan *Event, 1000),
		result:     make(chan *Result, 1000),
		internal:   mo.internal,
		replicator: mo.replicator,
	}
	return em
}

// Sender send events
type Sender interface {
	Send(et *Event)
}

// Manager manage events, including send events, handle events and handle result
type Manager interface {
	Sender

	HandleEvent()
	HandleResult()
}

type eventManager struct {
	events chan *Event

	internal time.Duration
	ticker   *time.Ticker

	cache  sync.Map
	result chan *Result

	replicator replicator.Replicator
}

func (e *eventManager) Send(et *Event) {
	if et.Result == nil {
		et.Result = e.result
		e.cache.Store(et.Id, et)
	}

	e.events <- et
}

func (e *eventManager) HandleResult() {
	gopool.Go(func(ctx context.Context) {
		e.resultHandle(ctx)
	})
}

func (e *eventManager) resultHandle(ctx context.Context) {
	for {
		select {
		case res, ok := <-e.result:
			if !ok {
				continue
			}
			if res.Error != nil {
				log.Error("result is error ", res.Error)
				continue
			}

			id := res.ID
			et, ok := e.cache.LoadAndDelete(id)
			if !ok {
				log.Warn(fmt.Sprintf("%s event not exist", id))
				continue
			}

			r, result := resource.New(et.(*Event).Event)
			if result != nil {
				log.Warn(fmt.Sprintf("new resource failed, %s", result.Message))
				continue
			}

			toSendEvent, err := r.FailHandle(ctx, res.Data.Code)
			if err != nil {
				continue
			}
			if toSendEvent != nil {
				e.Send(&Event{
					Event: toSendEvent,
				})
			}
		case <-ctx.Done():
			log.Info("result handle worker is closed")
			return
		}
	}
}

func (e *eventManager) Close() {
	e.ticker.Stop()
	close(e.result)
}

type syncEvents []*Event

func (s syncEvents) Len() int {
	return len(s)
}

func (s syncEvents) Less(i, j int) bool {
	return s[i].Timestamp < s[j].Timestamp
}

func (s syncEvents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (e *eventManager) HandleEvent() {
	gopool.Go(func(ctx context.Context) {
		e.handleEvent(ctx)
	})
}

func (e *eventManager) handleEvent(ctx context.Context) {
	events := make([]*Event, 0, 100)
	e.ticker = time.NewTicker(e.internal)
	for {
		select {
		case <-e.ticker.C:
			if len(events) == 0 {
				continue
			}
			send := events[:]

			events = make([]*Event, 0, 100)
			go e.handle(ctx, send)
		case event, ok := <-e.events:
			if !ok {
				return
			}

			events = append(events, event)
			if len(events) > 50 {
				send := events[:]
				events = make([]*Event, 0, 100)
				go e.handle(ctx, send)
			}
		case <-ctx.Done():
			e.Close()
			return
		}
	}
}

func (e *eventManager) handle(ctx context.Context, es syncEvents) {
	sort.Sort(es)

	sendEvents := make([]*v1sync.Event, 0, len(es))
	for _, event := range es {
		sendEvents = append(sendEvents, event.Event)
	}

	result, err := e.replicator.Replicate(ctx, &v1sync.EventList{
		Events: sendEvents,
	})

	if err != nil {
		log.Error("replicate failed", err)
		result = &v1sync.Results{
			Results: make(map[string]*v1sync.Result),
		}
	}

	for _, e := range es {
		e.Result <- &Result{
			ID:    e.Id,
			Data:  result.Results[e.Id],
			Error: err,
		}
	}
}

// Send sends event to replicator
func Send(e *Event) {
	log.Info(fmt.Sprintf("send event %s", e.Subject))
	m.Send(e)
}
