// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package alarm

import (
	nf "github.com/apache/servicecomb-service-center/pkg/notify"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/notify"
	"sync"
)

var (
	center *Center
	once   sync.Once
)

type Center struct {
	nf.Subscriber
	alarms util.ConcurrentMap
}

func (ac *Center) Alarm(id ID, fields ...Field) error {
	ae := &AlarmEvent{
		Event:  nf.NewEvent(nf.NOTIFTY, Subject, ""),
		Status: TypeActivated,
		Id:     id,
		Fields: util.NewJSONObject(),
	}
	for _, f := range fields {
		ae.Fields[f.Key] = f.Value
	}
	return notify.NotifyCenter().Publish(ae)
}

func (ac *Center) Clear(id ID, fields ...Field) error {
	ae := &AlarmEvent{
		Event:  nf.NewEvent(nf.NOTIFTY, Subject, ""),
		Status: TypeCleared,
		Id:     id,
		Fields: util.NewJSONObject(),
	}
	for _, f := range fields {
		ae.Fields[f.Key] = f.Value
	}
	return notify.NotifyCenter().Publish(ae)
}

func (ac *Center) AlarmList() (ls []*AlarmEvent) {
	ac.alarms.ForEach(func(item util.MapItem) (next bool) {
		ls = append(ls, item.Value.(*AlarmEvent))
		return true
	})
	return
}

func (ac *Center) ClearAll() {
	ac.alarms = util.ConcurrentMap{}
	return
}

func (ac *Center) OnMessage(evt nf.Event) {
	alarm := evt.(*AlarmEvent)
	ac.alarms.Put(alarm.Id, alarm)
}

func NewAlarmCenter() *Center {
	c := &Center{
		Subscriber: nf.NewSubscriber(nf.NOTIFTY, Subject, Group),
	}
	notify.NotifyCenter().AddSubscriber(c)
	return c
}

func AlarmCenter() *Center {
	once.Do(func() {
		center = NewAlarmCenter()
	})
	return center
}
