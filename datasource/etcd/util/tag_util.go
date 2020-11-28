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

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/kv"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func AddTagIntoETCD(ctx context.Context, domainProject string, serviceID string, dataTags map[string]string) *discovery.Error {
	key := path.GenerateServiceTagKey(domainProject, serviceID)
	data, err := json.Marshal(dataTags)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}

	resp, err := client.Instance().TxnWithCmp(ctx,
		[]client.PluginOp{client.OpPut(client.WithStrKey(key), client.WithValue(data))},
		[]client.CompareOp{client.OpCmp(
			client.CmpVer(util.StringToBytesWithNoCopy(path.GenerateServiceKey(domainProject, serviceID))),
			client.CmpNotEqual, 0)},
		nil)
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
	opts := append(FromContext(ctx), client.WithStrKey(key))
	resp, err := kv.Store().ServiceTag().Search(ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service[%s] tags file failed", serviceID)
		return tags, err
	}

	if len(resp.Kvs) != 0 {
		tags = resp.Kvs[0].Value.(map[string]string)
	}
	return tags, nil
}
