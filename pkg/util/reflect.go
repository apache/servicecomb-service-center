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
package util

import (
	"reflect"
	"sync"
	"unsafe"
)

var (
	reflector      *Reflector
	sliceTypeSize  = uint64(reflect.TypeOf(reflect.SliceHeader{}).Size())
	stringTypeSize = uint64(reflect.TypeOf(reflect.StringHeader{}).Size())
)

func init() {
	reflector = &Reflector{
		types: make(map[*uintptr]reflectObject, 100),
	}
}

type reflectObject struct {
	Type   reflect.Type
	Fields []reflect.StructField
}

type Reflector struct {
	types map[*uintptr]reflectObject
	mux   sync.RWMutex
}

func (r *Reflector) Load(obj interface{}) reflectObject {
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

	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		r.mux.Unlock()
		return reflectObject{}
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			r.mux.Unlock()
			return reflectObject{}
		}
		fallthrough
	case reflect.Interface:
		r.mux.Unlock()
		return r.Load(v.Elem().Interface())
	default:
		t = reflectObject{
			Type: reflect.TypeOf(obj),
		}
	}

	if v.Kind() != reflect.Struct {
		r.mux.Unlock()
		return t
	}

	fl := t.Type.NumField()
	if fl > 0 {
		t.Fields = make([]reflect.StructField, fl)
		for i := 0; i < fl; i++ {
			f := t.Type.Field(i)
			t.Fields[i] = f
		}
	}
	r.types[itab] = t
	r.mux.Unlock()
	return t
}

func ReflectObject(obj interface{}) (s reflectObject) {
	return reflector.Load(obj)
}

func Sizeof(obj interface{}) uint64 {
	selfRecurseMap := make(map[uintptr]struct{})
	return sizeof(reflect.ValueOf(obj), selfRecurseMap)
}

func sizeof(v reflect.Value, selfRecurseMap map[uintptr]struct{}) (s uint64) {
	if !v.IsValid() {
		return
	}

	if v.CanAddr() {
		selfRecurseMap[v.Addr().Pointer()] = struct{}{}
	}

	t := v.Type()
	s += uint64(t.Size())
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			break
		}
		if _, ok := selfRecurseMap[v.Pointer()]; ok {
			break
		}
		fallthrough
	case reflect.Interface:
		s += sizeof(v.Elem(), selfRecurseMap)
	case reflect.Struct:
		s -= uint64(t.Size())
		for i := 0; i < v.NumField(); i++ {
			s += sizeof(v.Field(i), selfRecurseMap)
		}
	case reflect.Array:
		if isValueType(t.Elem().Kind()) {
			break
		}
		s -= uint64(t.Size())
		for i := 0; i < v.Len(); i++ {
			s += sizeof(v.Index(i), selfRecurseMap)
		}
	case reflect.Slice:
		et := t.Elem()
		if isValueType(et.Kind()) {
			s += uint64(v.Len()) * uint64(et.Size())
			break
		}
		for i := 0; i < v.Len(); i++ {
			s += sizeof(v.Index(i), selfRecurseMap)
		}
	case reflect.Map:
		if v.IsNil() {
			break
		}
		kt, vt := t.Key(), t.Elem()
		if isValueType(kt.Kind()) && isValueType(vt.Kind()) {
			s += uint64(kt.Size()+vt.Size()) * uint64(v.Len())
			break
		}
		for _, k := range v.MapKeys() {
			s += sizeof(k, selfRecurseMap)
			s += sizeof(v.MapIndex(k), selfRecurseMap)
		}
	case reflect.String:
		s += uint64(v.Len())
	}
	return
}

func isValueType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}
