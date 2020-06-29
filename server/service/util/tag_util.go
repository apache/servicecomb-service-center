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
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

func AddTagIntoETCD(ctx context.Context, domainProject string, serviceId string, dataTags map[string]string) *scerr.Error {
	key := apt.GenerateServiceTagKey(domainProject, serviceId)
	data, err := json.Marshal(dataTags)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := backend.Registry().TxnWithCmp(ctx,
		[]registry.PluginOp{registry.OpPut(registry.WithStrKey(key), registry.WithValue(data))},
		[]registry.CompareOp{registry.OpCmp(
			registry.CmpVer(util.StringToBytesWithNoCopy(apt.GenerateServiceKey(domainProject, serviceId))),
			registry.CMP_NOT_EQUAL, 0)},
		nil)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
	}
	if !resp.Succeeded {
		return scerr.NewError(scerr.ErrServiceNotExists, "Service does not exist.")
	}
	return nil
}

func GetTagsUtils(ctx context.Context, domainProject, serviceId string) (tags map[string]string, err error) {
	key := apt.GenerateServiceTagKey(domainProject, serviceId)
	opts := append(FromContext(ctx), registry.WithStrKey(key))
	resp, err := backend.Store().ServiceTag().Search(ctx, opts...)
	if err != nil {
		log.Errorf(err, "get service[%s] tags file failed", serviceId)
		return tags, err
	}

	if len(resp.Kvs) != 0 {
		tags = resp.Kvs[0].Value.(map[string]string)
	}
	return tags, nil
}
