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
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s:json marshal tag data failed.", serviceId)
		return err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s: commit tag data into etcd failed.", serviceId)
		return err
	}
	return nil
}

func GetTagsUtils(ctx context.Context, tenant string, serviceId string) (map[string]string, error) {
	tags := map[string]string{}

	key := apt.GenerateServiceTagKey(tenant, serviceId)
	resp, err := store.Store().ServiceTag().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		return tags, err
	}
	if len(resp.Kvs) != 0 {
		util.LOGGER.Debugf("start unmarshal service tags file: %s", key)
		err = json.Unmarshal(resp.Kvs[0].Value, &tags)
		if err != nil {
			return tags, err
		}
	}
	return tags, nil
}
