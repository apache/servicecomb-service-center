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
	nf "github.com/apache/servicecomb-service-center/server/service/notification"
	"sync/atomic"
)

type ID int32

const (
	CodeBackendConnectionRefuse ID = iota
	CodeServerOverload
	CodeServiceQuotaLimit
	CodeInstanceQuotaLimit
	CodeDiagnoseFailure
	CodeInternalError
	typeEnd
)

const Subject = "__ALARM_CENTER__"

var latestId = int32(typeEnd)

type AlarmEvent struct {
	nf.Event
	Id     ID
	fields map[string]interface{}
}

func (ae *AlarmEvent) FieldBool(key string) bool {
	v, _ := ae.fields[key].(bool)
	return v
}

func (ae *AlarmEvent) FieldString(key string) string {
	v, _ := ae.fields[key].(string)
	return v
}

func (ae *AlarmEvent) FieldInt64(key string) int64 {
	v, _ := ae.fields[key].(int64)
	return v
}

func (ae *AlarmEvent) FieldInt(key string) int {
	v, _ := ae.fields[key].(int)
	return v
}

func (ae *AlarmEvent) FieldFloat64(key string) float64 {
	v, _ := ae.fields[key].(float64)
	return v
}

func RegisterAlarmSource() ID {
	return ID(atomic.AddInt32(&latestId, 1))
}

func Alarm(id ID, fields ...Field) error {
	ae := &AlarmEvent{
		Event:  nf.NewEvent(nf.NOTIFTY, Subject, ""),
		Id:     id,
		fields: make(map[string]interface{}, len(fields)),
	}
	for _, f := range fields {
		ae.fields[f.Key] = f.Value
	}
	return nf.GetNotifyService().Publish(ae)
}
