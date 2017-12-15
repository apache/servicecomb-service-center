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

func (r *Reflector) Load(obj interface{}) *StructType {
	r.mux.RLock()
	itab := *(**uintptr)(unsafe.Pointer(&obj))
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
		Type: reflect.TypeOf(obj),
	}

	fl := t.Type.NumField()
	if fl > 0 {
		t.Fields = make([]*reflect.StructField, fl)
		for i := 0; i < fl; i++ {
			f := t.Type.Field(i)
			t.Fields[i] = &f
		}
	}
	r.types[itab] = t
	r.mux.Unlock()
	return t
}

func LoadStruct(obj interface{}) *StructType {
	return reflector.Load(obj)
}
