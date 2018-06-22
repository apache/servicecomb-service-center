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
package backend

import (
	"errors"
	"fmt"
)

type Entity interface {
	Name() string
	Config() *Config
}

type entity struct {
	name string
	cfg  *Config
}

func (e *entity) Name() string {
	return e.name
}

func (e *entity) Config() *Config {
	return e.cfg
}

func InstallType(e Entity) (id StoreType, err error) {
	if e == nil {
		return NONEXIST, errors.New("invalid parameter")
	}
	for _, n := range TypeNames {
		if n == e.Name() {
			return NONEXIST, fmt.Errorf("redeclare store type '%s'", n)
		}
	}
	for _, r := range TypeConfig {
		if r.Prefix == e.Config().Prefix {
			return NONEXIST, fmt.Errorf("redeclare store root '%s'", r)
		}
	}

	id = StoreType(len(TypeNames))
	TypeNames = append(TypeNames, e.Name())
	TypeConfig[id] = e.Config()
	EventProxies[id] = NewEventProxy()
	return
}

func NewEntity(name string, cfg *Config) Entity {
	return &entity{
		name: name,
		cfg:  cfg,
	}
}
