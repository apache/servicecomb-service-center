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
package buildin

import (
	"context"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

var (
	closeCh    = make(chan struct{})
	noResponse = &registry.PluginResponse{}
)

func init() {
	close(closeCh)
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.REGISTRY, Name: "buildin", New: NewRegistry})
}

type Registry struct {
	ready chan int
}

func (ec *Registry) Err() (err <-chan error) {
	return
}
func (ec *Registry) Ready() <-chan struct{} {
	return closeCh
}
func (ec *Registry) PutNoOverride(ctx context.Context, opts ...registry.PluginOpOption) (bool, error) {
	return false, nil
}
func (ec *Registry) Do(ctx context.Context, opts ...registry.PluginOpOption) (*registry.PluginResponse, error) {
	return noResponse, nil
}
func (ec *Registry) Txn(ctx context.Context, ops []registry.PluginOp) (*registry.PluginResponse, error) {
	return noResponse, nil
}
func (ec *Registry) TxnWithCmp(ctx context.Context, success []registry.PluginOp, cmp []registry.CompareOp, fail []registry.PluginOp) (*registry.PluginResponse, error) {
	return noResponse, nil
}
func (ec *Registry) LeaseGrant(ctx context.Context, TTL int64) (leaseID int64, err error) {
	return 0, nil
}
func (ec *Registry) LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error) {
	return 0, nil
}
func (ec *Registry) LeaseRevoke(ctx context.Context, leaseID int64) error {
	return nil
}
func (ec *Registry) Watch(ctx context.Context, opts ...registry.PluginOpOption) error {
	return nil
}
func (ec *Registry) Compact(ctx context.Context, reserve int64) error {
	return nil
}
func (ec *Registry) Close() {
}

func NewRegistry() mgr.Instance {
	return &Registry{
		ready: make(chan int),
	}
}
