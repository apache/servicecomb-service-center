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
package mux

import (
	"github.com/ServiceComb/service-center/pkg/etcdsync"
	"reflect"
	"unsafe"
)

type MuxType string

func (m *MuxType) String() (s string) {
	pMT := (*reflect.StringHeader)(unsafe.Pointer(m))
	pStr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pStr.Data = pMT.Data
	pStr.Len = pMT.Len
	return
}

const (
	GLOBAL_LOCK  MuxType = "/global"
	PROCESS_LOCK MuxType = "/servicecenter"
)

func Lock(t MuxType) (*etcdsync.Locker, error) {
	return etcdsync.Lock(t.String())
}
