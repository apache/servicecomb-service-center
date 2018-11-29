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
	"testing"
	"time"
)

func TestGetNotifyService(t *testing.T) {
	notifyService := NewNotifyService()
	if notifyService == nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	if notifyService.Closed() != true {
		t.Fatalf("TestGetNotifyService failed")
	}

	err := notifyService.AddSubscriber(nil)
	if err == nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	err = notifyService.Publish(nil)
	if err == nil {
		t.Fatalf("TestGetNotifyService failed")
	}

	INSTANCE := RegisterType("INSTANCE", 1)
	notifyService.Start()
	notifyService.Start()
	if notifyService.Closed() != false {
		t.Fatalf("TestGetNotifyService failed")
	}
	select {
	case <-notifyService.Err():
		t.Fatalf("TestGetNotifyService failed")
	default:
	}

	s := NewSubscriber(-1, "s", "g")
	err = notifyService.AddSubscriber(s)
	if err == nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	notifyService.RemoveSubscriber(s)

	s = NewSubscriber(INSTANCE, "s", "g")
	err = notifyService.AddSubscriber(s)
	if err != nil {
		t.Fatalf("TestGetNotifyService failed, %v", err)
	}
	j := &baseEvent{INSTANCE, "s", "g"}
	err = notifyService.Publish(j)
	if err != nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	err = notifyService.Publish(NewNotifyServiceHealthCheckJob(NewNotifyServiceHealthChecker()))
	if err != nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	err = notifyService.Publish(NewNotifyServiceHealthCheckJob(s))
	if err != nil {
		t.Fatalf("TestGetNotifyService failed")
	}
	<-time.After(time.Second)
	notifyService.Stop()
	if notifyService.Closed() != true {
		t.Fatalf("TestGetNotifyService failed")
	}
}
