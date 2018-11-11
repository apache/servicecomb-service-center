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
	"github.com/apache/servicecomb-service-center/pkg/util"
)

type Subject struct {
	name   string
	groups *util.ConcurrentMap
}

func (s *Subject) Name() string {
	return s.name
}

func (s *Subject) Notify(job NotifyJob) {
	if len(job.Group()) == 0 {
		s.groups.ForEach(func(item util.MapItem) (next bool) {
			item.Value.(*Group).Notify(job)
			return true
		})
		return
	}

	itf, ok := s.groups.Get(job.Group())
	if !ok {
		return
	}
	itf.(*Group).Notify(job)
}

func (s *Subject) Groups(name string) *Group {
	g, ok := s.groups.Get(name)
	if !ok {
		return nil
	}
	return g.(*Group)
}

func (s *Subject) GetOrNewGroup(name string) *Group {
	item, _ := s.groups.Fetch(name, func() (interface{}, error) {
		return NewGroup(name), nil
	})
	return item.(*Group)
}

func (s *Subject) Remove(name string) {
	s.groups.Remove(name)
}

func (s *Subject) Size() int {
	return s.groups.Size()
}

func NewSubject(name string) *Subject {
	return &Subject{
		name:   name,
		groups: util.NewConcurrentMap(0),
	}
}
