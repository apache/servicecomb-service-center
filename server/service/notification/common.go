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
	"fmt"
	"strconv"
	"time"
)

const (
	DEFAULT_MAX_QUEUE          = 1000
	DEFAULT_INIT_SUBSCRIBERS   = 1000
	DEFAULT_ON_MESSAGE_TIMEOUT = 100 * time.Millisecond
	DEFAULT_TIMEOUT            = 30 * time.Second

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

type NotifyServiceConfig struct {
	AddTimeout    time.Duration
	NotifyTimeout time.Duration
	MaxQueue      int64
}

func (nsc NotifyServiceConfig) String() string {
	return fmt.Sprintf("{acceptQueue: %d, accept: %s, notify: %s}",
		nsc.MaxQueue, nsc.AddTimeout, nsc.NotifyTimeout)
}
