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

package notify

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/queue"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Processor struct {
	*queue.TaskQueue

	name     string
	subjects *util.ConcurrentMap
}

func (p *Processor) Name() string {
	return p.name
}

func (p *Processor) Accept(job Event) {
	p.Add(queue.Task{Object: job})
}

func (p *Processor) Handle(ctx context.Context, obj interface{}) {
	p.Notify(obj.(Event))
}

func (p *Processor) Notify(job Event) {
	if itf, ok := p.subjects.Get(job.Subject()); ok {
		itf.(*Subject).Notify(job)
	}
}

func (p *Processor) Subjects(name string) *Subject {
	itf, ok := p.subjects.Get(name)
	if !ok {
		return nil
	}
	return itf.(*Subject)
}

func (p *Processor) AddSubscriber(n Subscriber) {
	item, _ := p.subjects.Fetch(n.Subject(), func() (interface{}, error) {
		return NewSubject(n.Subject()), nil
	})
	item.(*Subject).GetOrNewGroup(n.Group()).AddSubscriber(n)
}

func (p *Processor) Remove(n Subscriber) {
	itf, ok := p.subjects.Get(n.Subject())
	if !ok {
		return
	}

	s := itf.(*Subject)
	g := s.Groups(n.Group())
	if g == nil {
		return
	}

	g.Remove(n.ID())

	if g.Size() == 0 {
		s.Remove(g.Name())
	}
	if s.Size() == 0 {
		p.subjects.Remove(s.Name())
	}
}

func (p *Processor) Clear() {
	p.subjects.Clear()
}

func NewProcessor(name string, queueSize int) *Processor {
	p := &Processor{
		TaskQueue: queue.NewTaskQueue(queueSize),
		name:      name,
		subjects:  util.NewConcurrentMap(0),
	}
	p.AddWorker(p)
	return p
}
