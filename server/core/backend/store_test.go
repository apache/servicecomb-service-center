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
	"context"
	"testing"
)

func TestStore(t *testing.T) {
	s := &KvStore{}
	s.Initialize()
	e := s.Entity(999)
	if e == nil {
		t.Fatalf("TestStore failed")
	}
	resp, err := e.Search(context.Background())
	if resp != nil || err != ErrNoImpl {
		t.Fatalf("TestStore failed")
	}
}

type extend struct {
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Config() *Config {
	return DefaultConfig().WithPrefix("/test")
}

func TestInstallType(t *testing.T) {
	s := &KvStore{}
	s.Initialize()
	id, err := s.Install(&extend{})
	if err != nil {
		t.Fatal(err)
	}
	if id == NOT_EXIST {
		t.Fatal(err)
	}
	if id.String() != "test" {
		t.Fatalf("TestInstallType failed")
	}

	id, err = s.Install(&extend{})
	if id != NOT_EXIST || err == nil {
		t.Fatal("installType fail", err)
	}
}
