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

	"github.com/apache/servicecomb-service-center/datasource/sdcommon"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/little-cui/etcdadpt"
)

type innerListWatch struct {
	Client etcdadpt.Client
	Prefix string

	rev int64
}

func (lw *innerListWatch) List(op sdcommon.ListWatchConfig) (*sdcommon.ListWatchResp, error) {
	otCtx, cancel := context.WithTimeout(op.Context, op.Timeout)
	defer cancel()
	resp, err := lw.Client.Do(otCtx, etcdadpt.WatchPrefixOpOptions(lw.Prefix)...)
	if err != nil {
		log.Error(fmt.Sprintf("list prefix %s failed, current rev: %d", lw.Prefix, lw.Revision()), err)
		return nil, err
	}
	lw.setRevision(resp.Revision)

	lwRsp := lw.doParsePluginRspToLwRsp(resp)

	return lwRsp, nil
}

func (lw *innerListWatch) Revision() int64 {
	return lw.rev
}

func (lw *innerListWatch) setRevision(rev int64) {
	lw.rev = rev
}

func (lw *innerListWatch) EventBus(op sdcommon.ListWatchConfig) *sdcommon.EventBus {
	return sdcommon.NewEventBus(lw, op)
}

func (lw *innerListWatch) DoWatch(ctx context.Context, f func(*sdcommon.ListWatchResp)) error {
	rev := lw.Revision()
	opts := append(
		etcdadpt.WatchPrefixOpOptions(lw.Prefix),
		etcdadpt.WithRev(rev+1),
		etcdadpt.WithWatchCallback(
			func(message string, resp *etcdadpt.Response) error {
				if resp == nil || len(resp.Kvs) == 0 {
					return fmt.Errorf("unknown event %s, watch prefix %s", resp, lw.Prefix)
				}

				lw.setRevision(resp.Revision)

				lwRsp := lw.doParsePluginRspToLwRsp(resp)
				switch resp.Action {
				case etcdadpt.ActionPut:
					lwRsp.Action = sdcommon.ActionPUT
				case etcdadpt.ActionDelete:
					lwRsp.Action = sdcommon.ActionDelete
				default:
					log.Warn(fmt.Sprintf("unrecognized action::%s", lwRsp.Action))
				}

				f(lwRsp)
				return nil
			}))

	err := lw.Client.Watch(ctx, opts...)
	if err != nil { // compact可能会导致watch失败 or message body size lager than 4MB
		log.Error(fmt.Sprintf("watch prefix %s failed, start rev: %d+1->%d->0", lw.Prefix, rev, lw.Revision()), err)

		lw.setRevision(0)
		f(nil)
	}
	return err
}

func (lw *innerListWatch) doParsePluginRspToLwRsp(pluginRsp *etcdadpt.Response) *sdcommon.ListWatchResp {
	lwRsp := &sdcommon.ListWatchResp{}

	lwRsp.Revision = pluginRsp.Revision

	for _, kv := range pluginRsp.Kvs {
		resource := sdcommon.Resource{}
		resource.Key = util.BytesToStringWithNoCopy(kv.Key)
		resource.ModRevision = kv.ModRevision
		resource.CreateRevision = kv.CreateRevision
		resource.Version = kv.Version
		resource.Value = kv.Value

		lwRsp.Resources = append(lwRsp.Resources, &resource)
	}
	return lwRsp
}
