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
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type innerListWatch struct {
	Client client.Registry
	Prefix string

	rev int64
}

func (lw *innerListWatch) List(op ListWatchConfig) (*client.PluginResponse, error) {
	otCtx, cancel := context.WithTimeout(op.Context, op.Timeout)
	defer cancel()
	resp, err := lw.Client.Do(otCtx, client.WatchPrefixOpOptions(lw.Prefix)...)
	if err != nil {
		log.Errorf(err, "list prefix %s failed, current rev: %d", lw.Prefix, lw.Revision())
		return nil, err
	}
	lw.setRevision(resp.Revision)
	return resp, nil
}

func (lw *innerListWatch) Revision() int64 {
	return lw.rev
}

func (lw *innerListWatch) setRevision(rev int64) {
	lw.rev = rev
}

func (lw *innerListWatch) Watch(op ListWatchConfig) Watcher {
	return newInnerWatcher(lw, op)
}

func (lw *innerListWatch) DoWatch(ctx context.Context, f func(*client.PluginResponse)) error {
	rev := lw.Revision()
	opts := append(
		client.WatchPrefixOpOptions(lw.Prefix),
		client.WithRev(rev+1),
		client.WithWatchCallback(
			func(message string, resp *client.PluginResponse) error {
				if resp == nil || len(resp.Kvs) == 0 {
					return fmt.Errorf("unknown event %s, watch prefix %s", resp, lw.Prefix)
				}

				lw.setRevision(resp.Revision)

				f(resp)
				return nil
			}))

	err := lw.Client.Watch(ctx, opts...)
	if err != nil { // compact可能会导致watch失败 or message body size lager than 4MB
		log.Errorf(err, "watch prefix %s failed, start rev: %d+1->%d->0", lw.Prefix, rev, lw.Revision())

		lw.setRevision(0)
		f(nil)
	}
	return err
}
