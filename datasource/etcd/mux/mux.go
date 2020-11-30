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

package mux

import (
	"reflect"
	"unsafe"

	"github.com/apache/servicecomb-service-center/pkg/etcdsync"
)

type Type string

func (m *Type) String() (s string) {
	pMT := (*reflect.StringHeader)(unsafe.Pointer(m))
	pStr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pStr.Data = pMT.Data
	pStr.Len = pMT.Len
	return
}

const (
	GlobalLock       Type = "/cse-sr/lock/global"
	DepQueueLock     Type = "/cse-sr/lock/dep-queue"
	ServiceClearLock Type = "/cse-sr/lock/service-clear"
)

func Lock(t Type) (*etcdsync.DLock, error) {
	return etcdsync.Lock(t.String(), -1, true)
}

func Try(t Type) (*etcdsync.DLock, error) {
	return etcdsync.Lock(t.String(), -1, false)
}
