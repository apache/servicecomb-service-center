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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
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

func (p *Processor) Add(n Subscriber) {
	itf, ok := p.subjects.Get(n.Subject())
	if !ok {
		itf = NewSubject(n.Subject())
		p.subjects.PutIfAbsent(n.Subject(), itf)
	}

	s := itf.(*Subject)
	g := s.Groups(n.Group())
	if g == nil {
		g = NewGroup(n.Group())
		s.Add(g)
	}

	g.Add(n)
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
	defer util.RecoverAndReport()
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

func NewProcessor(name string) *Processor {
	return &Processor{
		name:     name,
		subjects: util.NewConcurrentMap(0),
		queue:    make(chan NotifyJob, GetNotifyService().Config.MaxQueue),
	}
}
