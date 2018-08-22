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
	"github.com/apache/incubator-servicecomb-service-center/pkg/gopool"
	"testing"
)

func TestProcessor_Do(t *testing.T) {
	mock1 := &mockSubscriber{BaseSubscriber: NewSubscriber(INSTANCE, "s1", "g1")}
	mock2 := &mockSubscriber{BaseSubscriber: NewSubscriber(INSTANCE, "s1", "g2")}
	p := NewProcessor("p1", 0)
	gopool.Go(p.Do)
	if p.Name() != "p1" {
		t.Fatalf("TestProcessor_Do")
	}
	if p.Subjects(mock1.Subject()) != nil {
		t.Fatalf("TestProcessor_Do")
	}
	p.AddSubscriber(mock1)
	if p.Subjects(mock1.Subject()).Groups(mock1.Group()).Subscribers(mock1.Id()) != mock1 {
		t.Fatalf("TestProcessor_Do")
	}
	p.Remove(NewSubscriber(INSTANCE, "s2", "g1"))
	p.Remove(NewSubscriber(INSTANCE, "s1", "g2"))
	p.Remove(mock1)
	if p.Subjects(mock1.Subject()) != nil {
		t.Fatalf("TestProcessor_Do")
	}
	p.AddSubscriber(mock1)
	p.Clear()
	if p.Subjects(mock1.Subject()) != nil {
		t.Fatalf("TestProcessor_Do")
	}
	p.AddSubscriber(mock1)
	job := &BaseNotifyJob{group: "g1"}
	p.Accept(job)
	if mock1.job != nil {
		t.Fatalf("TestProcessor_Do")
	}
	job.subject = "s1"
	job.group = "g3"
	p.Accept(job)
	if mock1.job != nil {
		t.Fatalf("TestProcessor_Do")
	}
	job.subject = "s1"
	job.group = "g1"
	p.Accept(job)
	if mock1.job != job || mock2.job != nil {
		t.Fatalf("TestProcessor_Do")
	}
	job.subject = "s1"
	job.group = ""
	p.Accept(job)
	if mock1.job != job && mock2.job != job {
		t.Fatalf("TestProcessor_Do")
	}
}
