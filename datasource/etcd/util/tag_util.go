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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	esync "github.com/apache/servicecomb-service-center/datasource/etcd/sync"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

func AddTagIntoETCD(ctx context.Context, domainProject string, serviceID string, dataTags map[string]string) *errsvc.Error {
	key := path.GenerateServiceTagKey(domainProject, serviceID)
	data, err := json.Marshal(dataTags)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}

	opts := etcdadpt.Ops(etcdadpt.OpPut(etcdadpt.WithStrKey(key), etcdadpt.WithValue(data)))
	syncOpts, err := esync.GenUpdateOpts(ctx, datasource.ResourceKV, data, esync.WithOpts(map[string]string{"key": key}))
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	opts = append(opts, syncOpts...)
	resp, err := etcdadpt.TxnWithCmp(ctx, opts,
		etcdadpt.If(etcdadpt.NotEqualVer(path.GenerateServiceKey(domainProject, serviceID), 0)), nil)
	if err != nil {
		return discovery.NewError(discovery.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return discovery.NewError(discovery.ErrServiceNotExists, "Service does not exist.")
	}
	return nil
}

func GetTagsUtils(ctx context.Context, domainProject, serviceID string) (tags map[string]string, err error) {
	key := path.GenerateServiceTagKey(domainProject, serviceID)
	opts := append(FromContext(ctx), etcdadpt.WithStrKey(key))
	resp, err := sd.ServiceTag().Search(ctx, opts...)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s] tags file failed", serviceID), err)
		return tags, err
	}

	if len(resp.Kvs) != 0 {
		tags = resp.Kvs[0].Value.(map[string]string)
	}
	return tags, nil
}
