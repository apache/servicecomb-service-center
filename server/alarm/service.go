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
	"github.com/apache/servicecomb-service-center/pkg/log"
	nf "github.com/apache/servicecomb-service-center/pkg/notify"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm/model"
	"github.com/apache/servicecomb-service-center/server/notify"
	"sync"
)

var (
	service *AlarmService
	once    sync.Once
)

type AlarmService struct {
	nf.Subscriber
	alarms util.ConcurrentMap
}

func (ac *AlarmService) Raise(id model.ID, fields ...model.Field) error {
	ae := &model.AlarmEvent{
		Event:  nf.NewEvent(nf.NOTIFTY, Subject, ""),
		Status: Activated,
		Id:     id,
		Fields: util.NewJSONObject(),
	}
	for _, f := range fields {
		ae.Fields[f.Key] = f.Value
	}
	return notify.NotifyCenter().Publish(ae)
}

func (ac *AlarmService) Clear(id model.ID) error {
	ae := &model.AlarmEvent{
		Event:  nf.NewEvent(nf.NOTIFTY, Subject, ""),
		Status: Cleared,
		Id:     id,
	}
	return notify.NotifyCenter().Publish(ae)
}

func (ac *AlarmService) ListAll() (ls []*model.AlarmEvent) {
	ac.alarms.ForEach(func(item util.MapItem) (next bool) {
		ls = append(ls, item.Value.(*model.AlarmEvent))
		return true
	})
	return
}

func (ac *AlarmService) ClearAll() {
	ac.alarms = util.ConcurrentMap{}
	return
}

func (ac *AlarmService) OnMessage(evt nf.Event) {
	alarm := evt.(*model.AlarmEvent)
	switch alarm.Status {
	case Cleared:
		if itf, ok := ac.alarms.Get(alarm.Id); ok {
			if exist := itf.(*model.AlarmEvent); exist.Status != Cleared {
				exist.Status = Cleared
				alarm = exist
			}
		}
	default:
		ac.alarms.Put(alarm.Id, alarm)
	}
	log.Debugf("alarm[%s] %s, %v", alarm.Id, alarm.Status, alarm.Fields)
}

func NewAlarmService() *AlarmService {
	c := &AlarmService{
		Subscriber: nf.NewSubscriber(nf.NOTIFTY, Subject, Group),
	}
	notify.NotifyCenter().AddSubscriber(c)
	return c
}

func AlarmCenter() *AlarmService {
	once.Do(func() {
		service = NewAlarmService()
	})
	return service
}
