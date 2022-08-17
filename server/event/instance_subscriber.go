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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

var errBusy = errors.New("too busy")

type InstanceSubscriber struct {
	event.Subscriber
	Job chan *InstanceEvent
}

func (w *InstanceSubscriber) SetError(err error) {
	w.Subscriber.SetError(err)
	// 触发清理job
	e := w.Bus().Fire(event.NewUnhealthyEvent(w))
	if e != nil {
		log.Error("", e)
	}
}

func (w *InstanceSubscriber) OnAccept() {
	if w.Err() != nil {
		return
	}
	log.Debug(fmt.Sprintf("accepted by event service, %s watcher %s %s", w.Type(), w.Group(), w.Subject()))
}

// 被通知
func (w *InstanceSubscriber) OnMessage(evt event.Event) {
	if w.Err() != nil {
		return
	}

	wJob, ok := evt.(*InstanceEvent)
	if !ok {
		return
	}
	w.sendMessage(wJob)
}

func (w *InstanceSubscriber) sendMessage(evt *InstanceEvent) {
	defer log.Recover()

	metrics.ReportPendingCompleted(evt)

	select {
	case w.Job <- evt:
	default:
		log.Error(fmt.Sprintf("the %s watcher %s %s event queue is full, drop the blocked events",
			w.Type(), w.Group(), w.Subject()), nil)
		w.cleanup()
		w.Job <- evt
	}
}

func (w *InstanceSubscriber) cleanup() {
	for {
		select {
		case evt, ok := <-w.Job:
			if !ok {
				return
			}
			metrics.ReportPublishCompleted(evt, errBusy)
		default:
			return
		}
	}
}

func (w *InstanceSubscriber) Close() {
	w.cleanup()
	close(w.Job)
}

func NewInstanceSubscriber(serviceID, domainProject string) *InstanceSubscriber {
	watcher := &InstanceSubscriber{
		Subscriber: event.NewSubscriber(INSTANCE, domainProject, serviceID),
		Job:        make(chan *InstanceEvent, INSTANCE.QueueSize()),
	}
	return watcher
}
