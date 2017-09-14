//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package util

import (
	"encoding/json"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
)

func AddTagIntoETCD(ctx context.Context, tenant string, serviceId string, dataTags map[string]string) error {
	key := apt.GenerateServiceTagKey(tenant, serviceId)
	data, err := json.Marshal(dataTags)
	if err != nil {
		util.Logger().Errorf(err, "add tag into etcd,serviceId %s:json marshal tag data failed.", serviceId)
		return err
	}

	_, err = registry.GetRegisterCenter().Do(ctx,
		registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(err, "add tag into etcd,serviceId %s: commit tag data into etcd failed.", serviceId)
		return err
	}
	return nil
}

func GetTagsUtils(ctx context.Context, tenant, serviceId string, opts ...registry.PluginOpOption) (tags map[string]string, err error) {
	key := apt.GenerateServiceTagKey(tenant, serviceId)
	opts = append(opts, registry.WithStrKey(key))
	resp, err := store.Store().ServiceTag().Search(ctx, opts...)
	if err != nil {
		util.Logger().Errorf(err, "get service %s tags file failed", key)
		return tags, err
	}

	l := len(resp.Kvs)
	if l != 0 {
		tags = make(map[string]string, l)
		err = json.Unmarshal(resp.Kvs[0].Value, &tags)
		if err != nil {
			util.Logger().Errorf(err, "unmarshal service %s tags file failed", key)
			return nil, err
		}
	}
	return tags, nil
}
