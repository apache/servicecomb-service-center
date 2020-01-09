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
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

var (
	reflector *Reflector
	unknown   = new(reflectObject)
)

func init() {
	reflector = &Reflector{
		types: make(map[*uintptr]*reflectObject),
	}
}

type reflectObject struct {
	// full name
	FullName string
	Type     reflect.Type
	// if type is not struct, Fields is nil
	Fields []reflect.StructField
}

// Name returns a short name of the object type
func (o *reflectObject) Name() string {
	return FileLastName(o.FullName)
}

type Reflector struct {
	types map[*uintptr]*reflectObject
	mux   sync.RWMutex
}

func (r *Reflector) Load(obj interface{}) *reflectObject {
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

	t = new(reflectObject)
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		r.mux.Unlock()
		return unknown
	}
	switch v.Kind() {
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		r.mux.Unlock()
		if v.IsNil() {
			return unknown
		}
		e := v.Elem()
		if e.CanInterface() {
			return r.Load(e.Interface())
		}
		return unknown
	default:
		t.Type = reflect.TypeOf(obj)

		switch v.Kind() {
		case reflect.Func:
			r.mux.Unlock()
			f := runtime.FuncForPC(v.Pointer())
			if f == nil {
				return unknown
			}
			t.FullName = f.Name()
			return t
		case reflect.Struct:
			if fl := t.Type.NumField(); fl > 0 {
				t.Fields = make([]reflect.StructField, fl)
				for i := 0; i < fl; i++ {
					f := t.Type.Field(i)
					t.Fields[i] = f
				}
			}
			fallthrough
		default:
			t.FullName = t.Type.PkgPath() + "." + t.Type.Name()
		}
	}
	r.types[itab] = t
	r.mux.Unlock()
	return t
}

func Reflect(obj interface{}) *reflectObject {
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

func FormatFuncName(f string) string {
	i := strings.LastIndex(f, "/")
	j := strings.Index(f[i+1:], ".")
	if j < 1 {
		return "???"
	}
	_, fun := f[:i+j+1], f[i+j+2:]
	i = strings.LastIndex(fun, ".")
	return strings.TrimSuffix(fun[i+1:], "-fm") // trim the suffix of function closure name
}

func FuncName(f interface{}) string {
	return Reflect(f).Name()
}
