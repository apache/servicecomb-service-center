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
package kv

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
)

type extend struct {
	evts []sd.KvEvent
	cfg  *sd.Config
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Config() *sd.Config {
	return e.cfg
}

func TestInstallType(t *testing.T) {
	s := NewStore()

	// case: normal
	e := &extend{cfg: sd.Configure()}
	e.cfg.WithPrefix("/test").WithEventFunc(func(evt sd.KvEvent) {
		e.evts = append(e.evts, evt)
	})
	id, err := s.Install(e)
	if err != nil {
		t.Fatal(err)
	}
	if id == sd.TypeError {
		t.Fatal(err)
	}
	if id.String() != "test" {
		t.Fatalf("TestInstallType failed")
	}

	// case: inject config
	itf, _ := s.AddOns[id]
	cfg := itf.(AddOn).Config()
	if cfg == nil || cfg.OnEvent == nil {
		t.Fatal("installType fail", err)
	}

	cfg.OnEvent(sd.KvEvent{Revision: 1})
	if s.rev != 1 || len(e.evts) != 1 {
		t.Fatalf("TestInstallType failed")
	}

	// case: install again
	cfg = sd.Configure().WithPrefix("/test")
	id, err = s.Install(NewAddOn("test", cfg))
	if id != sd.TypeError || err == nil {
		t.Fatal("installType fail", err)
	}
}

func TestNewAddOn(t *testing.T) {
	s := NewStore()

	id, err := s.Install(NewAddOn("TestNewAddOn", nil))
	if id != sd.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(NewAddOn("", sd.Configure()))
	if id != sd.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(nil)
	if id != sd.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = s.Install(NewAddOn("TestNewAddOn", sd.Configure()))
	if id == sd.TypeError || err != nil {
		t.Fatalf("TestNewAddOn failed")
	}
	_, err = s.Install(NewAddOn("TestNewAddOn", sd.Configure()))
	if err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
}
