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
package store

import (
	"errors"
	"fmt"
)

type Entity interface {
	Name() string
	Prefix() string
	InitSize() int
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
	for _, r := range TypeRoots {
		if r == e.Prefix() {
			return NONEXIST, fmt.Errorf("redeclare store root '%s'", r)
		}
	}

	TypeNames = append(TypeNames, e.Name())
	id = StoreType(len(TypeNames) + 1) // +1 for typeEnd

	TypeRoots[id] = e.Prefix()
	TypeInitSize[id] = e.InitSize()

	EventProxies[id] = NewEventProxy()
	return
}
