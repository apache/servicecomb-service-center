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

//Package client is the abstraction of kv database operator
package client

import (
	"context"
	"errors"

	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
)

const (
	// the same as v3rpc.MaxOpsPerTxn = 128
	MaxTxnNumberOneTime = 128
)

var ErrNotUnique = errors.New("kv result is not unique")

// Registry is a abstraction of kv database operator
// Support etcd by default
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

//Put put kv
func Put(ctx context.Context, key string, value string) error {
	_, err := Instance().Do(ctx, PUT, WithStrKey(key), WithStrValue(value))
	return err
}

//Put put kv
func PutBytes(ctx context.Context, key string, value []byte) error {
	_, err := Instance().Do(ctx, PUT, WithStrKey(key), WithValue(value))
	return err
}

//Get get one kv
func Get(ctx context.Context, key string) (*mvccpb.KeyValue, error) {
	resp, err := Instance().Do(ctx, GET, WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if resp.Count != 1 {
		return nil, ErrNotUnique
	}
	return resp.Kvs[0], err
}

//List get kv list
func List(ctx context.Context, key string) ([]*mvccpb.KeyValue, int64, error) {
	resp, err := Instance().Do(ctx, GET, WithStrKey(key), WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	return resp.Kvs, resp.Count, nil
}

//Exist get one kv, if can not get return false
func Exist(ctx context.Context, key string) (bool, error) {
	resp, err := Instance().Do(ctx, GET, WithStrKey(key), WithCountOnly())
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}

func Delete(ctx context.Context, key string) (bool, error) {
	resp, err := Instance().Do(ctx, DEL, WithStrKey(key))
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}

func BatchCommit(ctx context.Context, opts []PluginOp) error {
	_, err := BatchCommitWithCmp(ctx, opts, nil, nil)
	return err
}

func BatchCommitWithCmp(ctx context.Context, opts []PluginOp,
	cmp []CompareOp, fail []PluginOp) (resp *PluginResponse, err error) {
	lenOpts := len(opts)
	tmpLen := lenOpts
	var tmpOpts []PluginOp
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MaxTxnNumberOneTime
		if tmpLen > 0 {
			tmpOpts = opts[i*MaxTxnNumberOneTime : (i+1)*MaxTxnNumberOneTime]
		} else {
			tmpOpts = opts[i*MaxTxnNumberOneTime : lenOpts]
		}
		resp, err = Instance().TxnWithCmp(ctx, tmpOpts, cmp, fail)
		if err != nil || !resp.Succeeded {
			return
		}
	}
	return
}

// KeepAlive will always return ok when cache is unavailable
// unless the cache response is LeaseNotFound
func KeepAlive(ctx context.Context, opts ...PluginOpOption) (int64, error) {
	op := OpPut(opts...)

	t := NewLeaseAsyncTask(op)
	if op.Mode == ModeNoCache {
		log.Debugf("keep alive lease WitchNoCache, request etcd server, op: %s", op)
		err := t.Do(ctx)
		ttl := t.TTL
		return ttl, err
	}

	err := task.GetService().Add(ctx, t)
	if err != nil {
		return 0, err
	}
	itf, err := task.GetService().LatestHandled(t.Key())
	if err != nil {
		return 0, err
	}
	pt := itf.(*LeaseTask)
	return pt.TTL, pt.Err()
}
