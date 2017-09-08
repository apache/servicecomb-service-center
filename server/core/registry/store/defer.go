//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package store

import "github.com/ServiceComb/service-center/util"

type DeferHandler interface {
	OnCondition(Cache, []*Event) bool
	HandleChan() <-chan *Event
}

type InstanceEventDeferHandler struct {
	events  []*Event
	deferCh chan *Event
}

func (iedh *InstanceEventDeferHandler) OnCondition(cache Cache, evts []*Event) bool {
	kvCache, ok := cache.(*KvCache)
	if !ok {
		return false
	}
	isDefer := len(evts)/kvCache.Size() >= 0
	if !isDefer {
		return false
	}
	if iedh.deferCh == nil {
		iedh.events = make([]*Event, len(evts))
		for i, evt := range evts {
			iedh.events[i] = evt
		}
		iedh.deferCh = make(chan *Event, event_block_size)
		util.Go(func(stopCh <-chan struct{}) {

		})
	}
	return true
}

func (iedh *InstanceEventDeferHandler) HandleChan() <-chan *Event {
	return iedh.deferCh
}
