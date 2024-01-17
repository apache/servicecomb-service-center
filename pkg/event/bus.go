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

	"github.com/apache/servicecomb-service-center/pkg/queue"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// Bus can fire the event aync and dispatch events to subscriber according to subject
type Bus struct {
	*queue.TaskQueue

	name     string
	subjects *util.ConcurrentMap
}

func (bus *Bus) Name() string {
	return bus.name
}

func (bus *Bus) Fire(evt Event) {
	// TODO add option if queue is full
	bus.Add(queue.Task{Payload: evt})
}

func (bus *Bus) Handle(_ context.Context, payload interface{}) {
	bus.fireAtOnce(payload.(Event))
}

func (bus *Bus) fireAtOnce(evt Event) {
	if itf, ok := bus.subjects.Get(evt.Subject()); ok {
		itf.(*Poster).Post(evt)
	} // else the evt will be discard
}

func (bus *Bus) Subjects(name string) *Poster {
	itf, ok := bus.subjects.Get(name)
	if !ok {
		return nil
	}
	return itf.(*Poster)
}

func (bus *Bus) AddSubscriber(n Subscriber) {
	item, _ := bus.subjects.Fetch(n.Subject(), func() (interface{}, error) {
		return NewPoster(n.Subject()), nil
	})
	item.(*Poster).GetOrNewGroup(n.Group()).AddMember(n)
}

func (bus *Bus) RemoveSubscriber(n Subscriber) {
	itf, ok := bus.subjects.Get(n.Subject())
	if !ok {
		return
	}

	s := itf.(*Poster)
	g := s.Groups(n.Group())
	if g == nil {
		return
	}

	g.RemoveMember(n.ID())

	if g.Size() == 0 {
		s.RemoveGroup(g.Name())
	}
	if s.Size() == 0 {
		bus.subjects.Remove(s.Subject())
	}
}

func (bus *Bus) Clear() {
	bus.subjects.Clear()
}

func NewBus(name string, queueSize int) *Bus {
	p := &Bus{
		TaskQueue: queue.NewTaskQueue(queueSize),
		name:      name,
		subjects:  util.NewConcurrentMap(0),
	}
	p.AddWorker(p)
	return p
}
