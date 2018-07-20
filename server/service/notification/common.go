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
	"strconv"
	"time"
)

const (
	DEFAULT_MAX_QUEUE          = 1000
	DEFAULT_ADD_JOB_TIMEOUT    = 1 * time.Second
	DEFAULT_SEND_TIMEOUT       = 5 * time.Second
	DEFAULT_HEARTBEAT_INTERVAL = 30 * time.Second
)

const (
	NOTIFTY NotifyType = iota
	INSTANCE
	typeEnd
)

type NotifyType int

func (nt NotifyType) String() string {
	if int(nt) < len(notifyTypeNames) {
		return notifyTypeNames[nt]
	}
	return "NotifyType" + strconv.Itoa(int(nt))
}

func (nt NotifyType) QueueSize() (s int) {
	if int(nt) < len(notifyTypeQueues) {
		s = notifyTypeQueues[nt]
	}
	if s <= 0 {
		s = DEFAULT_MAX_QUEUE
	}
	return
}

var notifyTypeNames = []string{
	NOTIFTY:  "NOTIFTY",
	INSTANCE: "INSTANCE",
}

var notifyTypeQueues = []int{
	INSTANCE: 100 * 1000,
}
