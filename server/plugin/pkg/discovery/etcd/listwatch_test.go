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
package etcd

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/registry/buildin"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"testing"
	"time"
)

type mockRegistry struct {
	*buildin.BuildinRegistry
	Response *registry.PluginResponse
}

func (c *mockRegistry) Do(ctx context.Context, opts ...registry.PluginOpOption) (*registry.PluginResponse, error) {
	if c.Response == nil {
		return nil, fmt.Errorf("error")
	}
	return c.Response, nil
}

func (c *mockRegistry) Watch(ctx context.Context, opts ...registry.PluginOpOption) error {
	op := registry.OptionsToOp(opts...)
	if c.Response == nil {
		return fmt.Errorf("error")
	}
	resp := *c.Response
	if len(c.Response.Kvs) > 0 {
		resp.Revision = c.Response.Kvs[0].ModRevision
	}
	err := op.WatchCallback("ok", &resp)
	if err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func TestPrefixListWatch(t *testing.T) {
	lw := &innerListWatch{
		Client: &mockRegistry{},
		Prefix: "a",
		rev:    1,
	}
	resp, err := lw.List(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	if resp != nil || err == nil || lw.Revision() != 1 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w := lw.Watch(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.EventBus()
	if resp != nil || lw.Revision() != 0 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w.Stop()

	test := &registry.PluginResponse{
		Revision: 2,
	}
	lw = &innerListWatch{
		Client: &mockRegistry{Response: test},
		Prefix: "a",
		rev:    1,
	}
	resp, err = lw.List(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	if resp == nil || err != nil || lw.Revision() != 2 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w = lw.Watch(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.EventBus()
	if resp != nil || lw.Revision() != 0 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w.Stop()

	test = &registry.PluginResponse{
		Kvs:      []*mvccpb.KeyValue{{ModRevision: 3}},
		Revision: 4,
	}
	lw = &innerListWatch{
		Client: &mockRegistry{Response: test},
		Prefix: "a",
		rev:    1,
	}
	resp, err = lw.List(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	if resp == nil || err != nil || lw.Revision() != 4 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w = lw.Watch(ListWatchConfig{Timeout: time.Second, Context: context.Background()})
	resp = <-w.EventBus()
	if resp == nil || lw.Revision() != 3 {
		t.Fatalf("TestPrefixListWatch failed")
	}
	w.Stop()
}

func TestListWatchConfig_String(t *testing.T) {
	lw := ListWatchConfig{Timeout: time.Second, Context: context.Background()}
	if lw.String() != "{timeout: 1s}" {
		t.Fatalf("TestListWatchConfig_String failed")
	}
}
