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

package client

import (
	"context"
	"fmt"

	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/task"
)

// KeepAlive will always return ok when cache is unavailable
// unless the cache response is LeaseNotFound
func KeepAlive(ctx context.Context, key string, leaseID int64, opts ...etcdadpt.OpOption) (int64, error) {
	op := etcdadpt.OpPut(append(opts, etcdadpt.WithStrKey(key), etcdadpt.WithLease(leaseID))...)

	t := NewLeaseAsyncTask(op)
	if op.Mode == etcdadpt.ModeNoCache {
		log.Debug(fmt.Sprintf("keep alive lease WitchNoCache, request etcd server, op: %s", op))
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
