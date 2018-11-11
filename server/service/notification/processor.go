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
package notification

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
)

type Processor struct {
	name     string
	subjects *util.ConcurrentMap
	queue    chan NotifyJob
}

func (p *Processor) Name() string {
	return p.name
}

func (p *Processor) Notify(job NotifyJob) {
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

	g.Remove(n.Id())

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

func (p *Processor) Accept(job NotifyJob) {
	defer log.Recover()
	p.queue <- job
}

func (p *Processor) Do(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			p.Notify(job)
		}
	}
}

func NewProcessor(name string, queue int) *Processor {
	return &Processor{
		name:     name,
		subjects: util.NewConcurrentMap(0),
		queue:    make(chan NotifyJob, queue),
	}
}
