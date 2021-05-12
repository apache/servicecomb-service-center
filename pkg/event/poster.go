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
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// Poster post the events of specified subject to the subscribers
type Poster struct {
	subject string
	groups  *util.ConcurrentMap
}

func (s *Poster) Subject() string {
	return s.subject
}

func (s *Poster) Post(job Event) {
	f := func(g *Group) {
		g.ForEach(func(m Subscriber) {
			m.OnMessage(job)
		})
	}

	if len(job.Group()) == 0 {
		s.groups.ForEach(func(item util.MapItem) (next bool) {
			f(item.Value.(*Group))
			return true
		})
		return
	}

	itf, ok := s.groups.Get(job.Group())
	if !ok {
		return
	}
	f(itf.(*Group))
}

func (s *Poster) Groups(name string) *Group {
	g, ok := s.groups.Get(name)
	if !ok {
		return nil
	}
	return g.(*Group)
}

func (s *Poster) GetOrNewGroup(name string) *Group {
	item, _ := s.groups.Fetch(name, func() (interface{}, error) {
		return NewGroup(name), nil
	})
	return item.(*Group)
}

func (s *Poster) RemoveGroup(name string) {
	s.groups.Remove(name)
}

func (s *Poster) Size() int {
	return s.groups.Size()
}

func NewPoster(subject string) *Poster {
	return &Poster{
		subject: subject,
		groups:  util.NewConcurrentMap(0),
	}
}
