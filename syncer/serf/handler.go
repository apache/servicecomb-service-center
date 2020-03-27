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

package serf

import (
	"fmt"
	"sync"

	"github.com/hashicorp/serf/serf"
)

// EventHandler interface
type EventHandler interface {
	Handle(event serf.Event) bool
	String() string
}

type eventHandler struct {
	filter  EventFilter
	handler HandleFunc
}

// NewEventHandler returns serf event handler
func NewEventHandler(filter EventFilter, handler HandleFunc) EventHandler {
	return &eventHandler{
		filter:  filter,
		handler: handler,
	}
}

// Handle invoke event handler
func (h *eventHandler) Handle(event serf.Event) bool {
	return h.filter.Invoke(event, h.handler)
}

// String returns event handler string
func (h *eventHandler) String() string {
	return "handler: filter = " + h.filter.String()
}

type onceHandler struct {
	once    sync.Once
	readyCh chan struct{}
	EventHandler
}

func onceEventHandler(handler EventHandler) *onceHandler {
	return &onceHandler{
		EventHandler: handler,
		readyCh:      make(chan struct{}),
	}
}

// Handle invoke once event handler
func (h *onceHandler) Handle(event serf.Event) bool {
	done := h.EventHandler.Handle(event)
	if done {
		h.once.Do(func() {
			close(h.readyCh)
		})
	}
	return done
}

// Ready Returns a channel that will be closed when once handler is invoked
func (h *onceHandler) Ready() <-chan struct{} {
	return h.readyCh
}

// EventFilter interface
type EventFilter interface {
	Invoke(event serf.Event, handler HandleFunc) bool
	String() string
}

// UserEventFilter event filter of serf user event
func UserEventFilter(name string) EventFilter {
	return userFilter{name: name}
}

// QueryFilter event filter of serf query
func QueryFilter(name string) EventFilter {
	return queryFilter{name: name}
}

// MemberJoinFilter event filter of member join
func MemberJoinFilter() EventFilter {
	return memberFilter{kind: serf.EventMemberJoin}
}

// MemberLeaveFilter event filter of member leave
func MemberLeaveFilter() EventFilter {
	return memberFilter{kind: serf.EventMemberLeave}
}

// MemberFailedFilter event filter of member failed
func MemberFailedFilter() EventFilter {
	return memberFilter{kind: serf.EventMemberFailed}
}

// MemberUpdateFilter event filter of member update
func MemberUpdateFilter() EventFilter {
	return memberFilter{kind: serf.EventMemberUpdate}
}

// MemberReapFilter event filter of member reap
func MemberReapFilter() EventFilter {
	return memberFilter{kind: serf.EventMemberReap}
}

type userFilter struct {
	name string
}

// Invoke user filter handler
func (f userFilter) Invoke(event serf.Event, handler HandleFunc) bool {
	if event.EventType() != serf.EventUser {
		return false
	}
	user, ok := event.(serf.UserEvent)
	return ok && f.name == user.Name && handler(user.Payload)
}

// String returns user filter string
func (f userFilter) String() string {
	return fmt.Sprintf("event kind = %s, name = %s", serf.EventUser.String(), f.name)
}

type queryFilter struct {
	name string
}

// Invoke query filter handler
func (f queryFilter) Invoke(event serf.Event, handler HandleFunc) bool {
	if event.EventType() != serf.EventQuery {
		return false
	}
	query, ok := event.(*serf.Query)
	return ok && f.name == query.Name && handler(query.Payload)
}

// String returns query filter string
func (f queryFilter) String() string {
	return fmt.Sprintf("event kind = %s, name = %s", serf.EventQuery.String(), f.name)
}

type memberFilter struct {
	kind serf.EventType
}

// Invoke member filter handler
func (f memberFilter) Invoke(event serf.Event, handler HandleFunc) bool {
	return event.EventType() == f.kind && handler()
}

// String returns member filter string
func (f memberFilter) String() string {
	return fmt.Sprintf("event kind = %s", f.kind.String())
}
