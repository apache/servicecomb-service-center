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
package validate

import (
	"reflect"
	"sync"
	"unsafe"
)

var reflector *Reflector

func init() {
	reflector = &Reflector{
		types: make(map[*uintptr]*StructType, 100),
	}
}

type StructType struct {
	Type   reflect.Type
	Fields []*reflect.StructField
}

type Reflector struct {
	types map[*uintptr]*StructType
	mux   sync.RWMutex
}

func (r *Reflector) Load(d interface{}) *StructType {
	r.mux.RLock()
	itab := *(**uintptr)(unsafe.Pointer(&d))
	t, ok := r.types[itab]
	r.mux.RUnlock()
	if ok {
		return t
	}

	r.mux.Lock()
	t, ok = r.types[itab]
	if ok {
		r.mux.Unlock()
		return t
	}
	t = &StructType{
		Type: reflect.TypeOf(d),
	}
	t.Fields = make([]*reflect.StructField, t.Type.NumField())
	for i := 0; i < t.Type.NumField(); i++ {
		f := t.Type.Field(i)
		t.Fields[i] = &f
	}
	r.types[itab] = t
	r.mux.Unlock()
	return t
}

func LoadStruct(d interface{}) *StructType {
	return reflector.Load(d)
}
