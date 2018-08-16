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
package discovery

import (
	"errors"
	"fmt"
)

type AddOn interface {
	Name() string
	Config() *Config
}

type addOn struct {
	name string
	cfg  *Config
}

func (e *addOn) Name() string {
	return e.name
}

func (e *addOn) Config() *Config {
	return e.cfg
}

func NewAddOn(name string, cfg *Config) AddOn {
	return &addOn{
		name: name,
		cfg:  cfg,
	}
}

func Install(e AddOn) (id StoreType, err error) {
	if e == nil {
		return TypeError, errors.New("invalid parameter")
	}
	for _, n := range typeNames {
		if n == e.Name() {
			return TypeError, fmt.Errorf("redeclare store type '%s'", n)
		}
	}

	id = RegisterStoreType(e.Name())
	EventProxy(id).InjectConfig(e.Config())
	return
}
