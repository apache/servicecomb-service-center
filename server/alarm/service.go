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
	nf "github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm/model"
	"github.com/apache/servicecomb-service-center/server/event"
	"sync"
)

var (
	service *Service
	once    sync.Once
)

type Service struct {
	nf.Subscriber
	alarms util.ConcurrentMap
}

func (ac *Service) Raise(id model.ID, fields ...model.Field) error {
	ae := &model.AlarmEvent{
		Event:  nf.NewEvent(ALARM, Subject, ""),
		Status: Activated,
		ID:     id,
		Fields: util.NewJSONObject(),
	}
	for _, f := range fields {
		ae.Fields[f.Key] = f.Value
	}
	return event.Center().Fire(ae)
}

func (ac *Service) Clear(id model.ID) error {
	ae := &model.AlarmEvent{
		Event:  nf.NewEvent(ALARM, Subject, ""),
		Status: Cleared,
		ID:     id,
	}
	return event.Center().Fire(ae)
}

func (ac *Service) ListAll() (ls []*model.AlarmEvent) {
	ac.alarms.ForEach(func(item util.MapItem) (next bool) {
		ls = append(ls, item.Value.(*model.AlarmEvent))
		return true
	})
	return
}

func (ac *Service) ClearAll() {
	ac.alarms = util.ConcurrentMap{}
}

func (ac *Service) OnMessage(evt nf.Event) {
	alarm := evt.(*model.AlarmEvent)
	switch alarm.Status {
	case Cleared:
		if itf, ok := ac.alarms.Get(alarm.ID); ok {
			if exist := itf.(*model.AlarmEvent); exist.Status != Cleared {
				exist.Status = Cleared
				alarm = exist
			}
		}
	default:
		ac.alarms.Put(alarm.ID, alarm)
	}
	log.Debugf("alarm[%s] %s, %v", alarm.ID, alarm.Status, alarm.Fields)
}

func NewAlarmService() *Service {
	c := &Service{
		Subscriber: nf.NewSubscriber(ALARM, Subject, Group),
	}
	err := event.Center().AddSubscriber(c)
	if err != nil {
		log.Error("", err)
	}
	return c
}
