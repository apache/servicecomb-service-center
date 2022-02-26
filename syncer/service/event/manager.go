/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package event

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/metrics"
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
	CanNotAbandon bool

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

	if e.checkThreshold(et) {
		return
	}

	e.events <- et
}

func (e *eventManager) checkThreshold(et *Event) bool {
	metrics.PendingEventSet(int64(len(e.events)))
	if len(e.events) < cap(e.events) {
		return false
	}

	log.Warn(fmt.Sprintf("events reaches the limit %d", cap(e.events)))
	if et.CanNotAbandon {
		return false
	}

	log.Warn(fmt.Sprintf("drop event %s", et.Flag()))
	metrics.AbandonEventAdd()
	return true
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

			id := res.ID
			et, ok := e.cache.LoadAndDelete(id)
			if !ok {
				log.Warn(fmt.Sprintf("%s event not exist", id))
				continue
			}

			event := et.(*Event).Event
			r, result := resource.New(event)
			if result != nil {
				log.Warn(fmt.Sprintf("new resource failed, %s", result.Message))
				continue
			}

			if res.Error != nil {
				log.Error(fmt.Sprintf("result is error %s", event.Flag()), res.Error)
				if r.CanDrop() {
					log.Warn(fmt.Sprintf("drop event %s", event.Flag()))
					continue
				}
				log.Info(fmt.Sprintf("resend event %s", event.Flag()))
				e.Send(&Event{
					Event: event,
				})

				continue
			}

			toSendEvent, err := r.FailHandle(ctx, res.Data.Code)
			if err != nil {
				log.Warn(fmt.Sprintf("event %s fail handle failed, %s", event.Flag(), err.Error()))
				continue
			}
			if toSendEvent != nil {
				log.Info(fmt.Sprintf("resend event %s", toSendEvent.Flag()))
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
