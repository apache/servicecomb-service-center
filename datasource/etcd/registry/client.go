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

package registry

import (
	"context"
)

type Registry interface {
	Err() <-chan error
	Ready() <-chan struct{}
	PutNoOverride(ctx context.Context, opts ...PluginOpOption) (bool, error)
	Do(ctx context.Context, opts ...PluginOpOption) (*PluginResponse, error)
	Txn(ctx context.Context, ops []PluginOp) (*PluginResponse, error)
	TxnWithCmp(ctx context.Context, success []PluginOp, cmp []CompareOp, fail []PluginOp) (*PluginResponse, error)
	LeaseGrant(ctx context.Context, TTL int64) (leaseID int64, err error)
	LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error)
	LeaseRevoke(ctx context.Context, leaseID int64) error
	// this function block util:
	// 1. connection error
	// 2. call send function failed
	// 3. response.Err()
	// 4. time out to watch, but return nil
	Watch(ctx context.Context, opts ...PluginOpOption) error
	Compact(ctx context.Context, reserve int64) error
	Close()
}
