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
	"github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"time"
)

const AddJobTimeout = 1 * time.Second

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
	log.Debugf("accepted by event service, %s watcher %s %s", w.Type(), w.Group(), w.Subject())
}

//被通知
func (w *InstanceSubscriber) OnMessage(job event.Event) {
	if w.Err() != nil {
		return
	}

	wJob, ok := job.(*InstanceEvent)
	if !ok {
		return
	}
	w.sendMessage(wJob)
}

func (w *InstanceSubscriber) sendMessage(job *InstanceEvent) {
	defer log.Recover()
	select {
	case w.Job <- job:
	default:
		timer := time.NewTimer(w.Timeout())
		select {
		case w.Job <- job:
			timer.Stop()
		case <-timer.C:
			log.Errorf(nil,
				"the %s watcher %s %s event queue is full[over %s], drop the event %v",
				w.Type(), w.Group(), w.Subject(), w.Timeout(), job)
		}
	}
}

func (w *InstanceSubscriber) Timeout() time.Duration {
	return AddJobTimeout
}

func (w *InstanceSubscriber) Close() {
	close(w.Job)
}

func NewInstanceSubscriber(serviceID, domainProject string) *InstanceSubscriber {
	watcher := &InstanceSubscriber{
		Subscriber: event.NewSubscriber(INSTANCE, domainProject, serviceID),
		Job:        make(chan *InstanceEvent, INSTANCE.QueueSize()),
	}
	return watcher
}
