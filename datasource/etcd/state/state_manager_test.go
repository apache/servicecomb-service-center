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
package state_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/stretchr/testify/assert"
)

type extend struct {
	evts []kvstore.Event
	cfg  *kvstore.Options
}

func (e *extend) Name() string {
	return "test"
}

func (e *extend) Config() *kvstore.Options {
	return e.cfg
}

func TestRegister(t *testing.T) {
	s := &state.Manager{}
	s.Initialize()

	// case: normal
	e := &extend{cfg: kvstore.NewOptions()}
	e.cfg.WithPrefix("/test").WithEventFunc(func(evt kvstore.Event) {
		e.evts = append(e.evts, evt)
	})
	id, err := state.Register(e)
	assert.NoError(t, err)
	assert.NotEqual(t, kvstore.TypeError, id)
	assert.Equal(t, "test", id.String())

	// case: inject config
	itf, _ := state.Plugins()[id]
	cfg := itf.Config()
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.OnEvent)

	s.InjectConfig(cfg)

	cfg.OnEvent(kvstore.Event{Revision: 1})
	assert.Equal(t, int64(1), s.Rev)
	assert.Equal(t, 1, len(e.evts))

	// case: install again
	cfg = kvstore.NewOptions().WithPrefix("/test")
	id, err = state.Register(state.NewPlugin("test", cfg))
	assert.Equal(t, kvstore.TypeError, id)
	assert.Error(t, err)
}

func TestNewPlugin(t *testing.T) {
	s := &state.Manager{}
	s.Initialize()

	id, err := state.Register(state.NewPlugin("TestNewAddOn", nil))
	if id != kvstore.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = state.Register(state.NewPlugin("", kvstore.NewOptions()))
	if id != kvstore.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = state.Register(nil)
	if id != kvstore.TypeError || err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
	id, err = state.Register(state.NewPlugin("TestNewAddOn", kvstore.NewOptions()))
	if id == kvstore.TypeError || err != nil {
		t.Fatalf("TestNewAddOn failed")
	}
	_, err = state.Register(state.NewPlugin("TestNewAddOn", kvstore.NewOptions()))
	if err == nil {
		t.Fatalf("TestNewAddOn failed")
	}
}
