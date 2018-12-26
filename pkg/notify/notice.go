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

import "time"

type Event interface {
	Type() Type
	Subject() string // required!
	Group() string   // broadcast all the subscriber of the same subject if group is empty
	CreateAt() time.Time
}

type baseEvent struct {
	nType    Type
	subject  string
	group    string
	createAt time.Time
}

func (s *baseEvent) Type() Type {
	return s.nType
}

func (s *baseEvent) Subject() string {
	return s.subject
}

func (s *baseEvent) Group() string {
	return s.group
}

func (s *baseEvent) CreateAt() time.Time {
	return s.createAt
}

func NewEvent(t Type, s string, g string) Event {
	return &baseEvent{t, s, g, time.Now()}
}
