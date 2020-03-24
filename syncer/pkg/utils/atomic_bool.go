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

package utils

import (
	"sync"
	"sync/atomic"
)

// AtomicBool struct
type AtomicBool struct {
	m      sync.Mutex
	status uint32
}

// NewAtomicBool returns an atomic bool
func NewAtomicBool(b bool) *AtomicBool {
	var status uint32
	if b {
		status = 1
	}
	return &AtomicBool{status: status}
}

// Bool returns a bool value
func (a *AtomicBool) Bool() bool {
	return atomic.LoadUint32(&a.status)&1 == 1
}

// DoToReverse Do something and reverse the status
func (a *AtomicBool) DoToReverse(when bool, fn func()) {
	if a.Bool() != when {
		return
	}

	a.m.Lock()
	fn()
	atomic.StoreUint32(&a.status, a.status^1)
	a.m.Unlock()
}
