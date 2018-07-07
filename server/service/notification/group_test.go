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

import "testing"

type mockSubscriber struct {
	*BaseSubscriber
	job NotifyJob
}

func (s *mockSubscriber) OnMessage(job NotifyJob) {
	s.job = job
}

func TestGroup_Add(t *testing.T) {
	m := NewSubscriber(INSTANCE, "s1", "g1")
	g := NewGroup("g1")
	if g.Name() != "g1" {
		t.Fatalf("TestGroup_Add failed")
	}
	if g.AddSubscriber(m) != m {
		t.Fatalf("TestGroup_Add failed")
	}
	if g.AddSubscriber(NewSubscriber(INSTANCE, "s1", "g1")) == m {
		t.Fatalf("TestGroup_Add failed")
	}
	same := *m
	if g.AddSubscriber(&same) != m {
		t.Fatalf("TestGroup_Add failed")
	}
	if g.Size() != 2 {
		t.Fatalf("TestGroup_Add failed")
	}
	g.Remove(m.Id())
	if g.Size() != 1 {
		t.Fatalf("TestGroup_Add failed")
	}
	if g.Subscribers(m.Id()) == m {
		t.Fatalf("TestGroup_Add failed")
	}

	mock := &mockSubscriber{BaseSubscriber: NewSubscriber(INSTANCE, "s1", "g1")}
	if g.AddSubscriber(mock) != mock {
		t.Fatalf("TestGroup_Add failed")
	}
	if g.Subscribers(mock.Id()) != mock {
		t.Fatalf("TestGroup_Add failed")
	}
	job := &BaseNotifyJob{nType: INSTANCE}
	g.Notify(job)
	if mock.job != job {
		t.Fatalf("TestGroup_Add failed")
	}
}
