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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"unicode/utf8"
)

type ValidateRule struct {
	Min    int
	Max    int
	Length int
	Regexp *regexp.Regexp
}

func (v *ValidateRule) String() string {
	toString := "rule: {"
	if v.Min != 0 {
		toString = fmt.Sprintf("%s Min: %d,", toString, v.Min)
	}
	if v.Max != 0 {
		toString = fmt.Sprintf("%s Max: %d,", toString, v.Max)
	}
	if v.Length != 0 {
		toString = fmt.Sprintf("%s Length: %d,", toString, v.Length)
	}
	if v.Regexp != nil {
		toString = fmt.Sprintf("%s RegEx: %s }", toString, v.Regexp)
	}
	return toString
}

func (v *ValidateRule) Match(s interface{}) bool {
	var invalid bool = false
	sv := reflect.ValueOf(s)
	if v.Min > 0 && !invalid {
		switch sv.Kind() {
		case reflect.String:
			invalid = len(sv.String()) < v.Min
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			invalid = sv.Int() < int64(v.Min)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			invalid = sv.Uint() < uint64(v.Min)
		case reflect.Float32, reflect.Float64:
			invalid = sv.Float() < float64(v.Min)
		case reflect.Slice, reflect.Map, reflect.Array:
			invalid = sv.Len() < v.Min
		default:
			invalid = false
		}
	}
	if v.Max > 0 && !invalid {
		switch sv.Kind() {
		case reflect.String:
			invalid = utf8.RuneCountInString(sv.String()) > v.Max
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			invalid = sv.Int() > int64(v.Max)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			invalid = sv.Uint() > uint64(v.Max)
		case reflect.Float32, reflect.Float64:
			invalid = sv.Float() > float64(v.Max)
		case reflect.Slice, reflect.Map, reflect.Array:
			invalid = sv.Len() > v.Max
		default:
			invalid = false
		}
	}
	if v.Length > 0 && !invalid {
		switch sv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			invalid = sv.Len() > v.Length
		case reflect.String:
			invalid = utf8.RuneCountInString(sv.String()) > v.Length
		default:
			invalid = false
		}
	}
	if v.Regexp != nil && !invalid {
		switch sv.Kind() {
		case reflect.Map:
			itemV := &ValidateRule{
				Regexp: v.Regexp,
			}
			keys := sv.MapKeys()
			for _, key := range keys {
				if !itemV.Match(key.Interface()) {
					invalid = true
					break
				}
				if !itemV.Match(sv.MapIndex(key).Interface()) {
					invalid = true
					break
				}
			}
		case reflect.Slice, reflect.Array:
			itemV := &ValidateRule{
				Regexp: v.Regexp,
			}
			for i, l := 0, sv.Len(); i < l; i++ {
				if !itemV.Match(sv.Index(i).Interface()) {
					invalid = true
					break
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64:
			var str string
			switch sv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				str = strconv.FormatInt(sv.Int(), 10)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				str = strconv.FormatUint(sv.Uint(), 10)
			case reflect.Float32, reflect.Float64:
				str = strconv.FormatFloat(sv.Float(), 'f', -1, 64)
			}
			invalid = !v.Regexp.MatchString(str)
		default:
			str, ok := s.(string)
			if ok {
				invalid = !v.Regexp.MatchString(str)
			} else {
				invalid = true
			}
		}
	}
	return !invalid
}

type Validator struct {
	rules map[string](*ValidateRule)
	subs  map[string](*Validator)
}

func (v *Validator) GetRule(name string) *ValidateRule {
	if v.rules == nil {
		return nil
	}
	return v.rules[name]
}

func (v *Validator) AddRule(name string, rule *ValidateRule) {
	if v.rules == nil {
		v.rules = make(map[string](*ValidateRule))
	}
	v.rules[name] = rule
}

func (v *Validator) AddRules(in map[string](*ValidateRule)) {
	if len(in) == 0 {
		return
	}
	for key, value := range in {
		v.AddRule(key, value)
	}
}

func (v *Validator) AddSub(name string, s *Validator) {
	if v.subs == nil {
		v.subs = make(map[string](*Validator))
	}
	v.subs[name] = s
}

func (v *Validator) GetRules() map[string](*ValidateRule) {
	return v.rules
}

func (v *Validator) Validate(s interface{}) error {
	sv := reflect.ValueOf(s)
	st := reflect.TypeOf(s)
	if sv.Kind() == reflect.Ptr && !sv.IsNil() {
		return v.Validate(sv.Elem().Interface())
	}
	if sv.Kind() == reflect.Slice && !sv.IsNil() {
		for i, l := 0, sv.Len(); i < l; i++ {
			err := v.Validate(sv.Index(i).Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	if sv.Kind() != reflect.Struct {
		return errors.New("not support validate type")
	}

	for i, l := 0, sv.NumField(); i < l; i++ {
		field := sv.Field(i)
		fieldName := st.Field(i).Name
		validator, ok := v.subs[fieldName]
		if ok {
			if (field.Kind() != reflect.Ptr && field.Kind() != reflect.Slice) || field.IsNil() {
				continue
			}
			err := validator.Validate(field.Interface())
			if err != nil {
				return err
			}
		}
		validate, ok := v.rules[fieldName]
		if ok {
			fi := field.Interface()
			if field.Kind() == reflect.Ptr && !field.IsNil() {
				fi = field.Elem().Interface()
				fsv := reflect.ValueOf(fi)
				if fsv.Kind() == reflect.Struct {
					err := v.Validate(fi)
					if err != nil {
						return err
					}
					continue
				}
			}
			// TODO null pointer如何校验
			if field.Kind() != reflect.Ptr && !validate.Match(fi) {
				return fmt.Errorf("%s validate failed, %s", fieldName, validate)
			}
		}
	}
	return nil
}
