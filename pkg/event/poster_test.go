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

import "testing"

func TestSubject_Fetch(t *testing.T) {
	s := NewPoster("s1")
	if s.Subject() != "s1" {
		t.Fatalf("TestSubject_Fetch failed")
	}
	g := s.GetOrNewGroup("g1")
	if g == nil {
		t.Fatalf("TestSubject_Fetch failed")
	}
	if s.GetOrNewGroup(g.Name()) != g {
		t.Fatalf("TestSubject_Fetch failed")
	}
	o := s.GetOrNewGroup("g2")
	if s.Groups("g2") != o {
		t.Fatalf("TestSubject_Fetch failed")
	}
	if s.Size() != 2 {
		t.Fatalf("TestSubject_Fetch failed")
	}
	s.RemoveGroup(o.Name())
	if s.Groups("g2") != nil {
		t.Fatalf("TestSubject_Fetch failed")
	}
	if s.Size() != 1 {
		t.Fatalf("TestSubject_Fetch failed")
	}
	INSTANCE := RegisterType("INSTANCE", 1)
	mock1 := &mockSubscriber{Subscriber: NewSubscriber(INSTANCE, "s1", "g1")}
	mock2 := &mockSubscriber{Subscriber: NewSubscriber(INSTANCE, "s1", "g2")}
	g.AddMember(mock1)
	job := &baseEvent{group: "g3"}
	s.Post(job)
	if mock1.job != nil || mock2.job != nil {
		t.Fatalf("TestSubject_Fetch failed")
	}
	job.group = "g1"
	s.Post(job)
	if mock1.job != job || mock2.job != nil {
		t.Fatalf("TestSubject_Fetch failed")
	}
	job.group = ""
	s.Post(job)
	if mock1.job != job && mock2.job != job {
		t.Fatalf("TestSubject_Fetch failed")
	}
}
