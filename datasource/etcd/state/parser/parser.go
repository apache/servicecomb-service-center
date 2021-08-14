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

package parser

import (
	"encoding/json"
	"errors"

	"github.com/go-chassis/foundation/stringutil"
)

var (
	ErrParseNilPoint  = errors.New("parse nil point")
	ErrTargetNilPoint = errors.New("target is nil point")

	newBytes  CreateValueFunc = func() interface{} { return []byte(nil) }
	newString CreateValueFunc = func() interface{} { return "" }
	newMap    CreateValueFunc = func() interface{} { return make(map[string]string) }

	BytesParser  = New(newBytes, UnParse)
	StringParser = New(newString, TextUnmarshal)
	MapParser    = New(newMap, MapUnmarshal)

	UnParse ParseValueFunc = func(src []byte, dist interface{}) error {
		if dist == nil {
			return ErrTargetNilPoint
		}
		d := dist.(*interface{})
		*d = src
		return nil
	}
	TextUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		if dist == nil {
			return ErrTargetNilPoint
		}
		d := dist.(*interface{})
		*d = stringutil.Bytes2str(src)
		return nil
	}
	MapUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		if err := check(src, dist); err != nil {
			return err
		}
		d := dist.(*interface{})
		m := (*d).(map[string]string)
		return json.Unmarshal(src, &m)
	}
	JSONUnmarshal ParseValueFunc = func(src []byte, dist interface{}) error {
		if err := check(src, dist); err != nil {
			return err
		}
		d := dist.(*interface{})
		return json.Unmarshal(src, *d)
	}
)

// CreateValueFunc the construct func of value object
type CreateValueFunc func() interface{}

// ParseValueFunc the func to parse src to dist object
type ParseValueFunc func(src []byte, dist interface{}) error

type Parser interface {
	Unmarshal(src []byte) (interface{}, error)
}

type CommonParser struct {
	NewFunc  CreateValueFunc
	FromFunc ParseValueFunc
}

func (p *CommonParser) Unmarshal(src []byte) (interface{}, error) {
	v := p.NewFunc()
	if err := p.FromFunc(src, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func check(src []byte, dist interface{}) error {
	if src == nil {
		return ErrParseNilPoint
	}
	if dist == nil {
		return ErrTargetNilPoint
	}
	return nil
}

func New(valueFunc CreateValueFunc, parseValueFunc ParseValueFunc) Parser {
	return &CommonParser{NewFunc: valueFunc, FromFunc: parseValueFunc}
}
