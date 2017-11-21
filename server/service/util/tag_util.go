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
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/backend"
	"github.com/ServiceComb/service-center/server/core/backend/store"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"golang.org/x/net/context"
)

func AddTagIntoETCD(ctx context.Context, domainProject string, serviceId string, dataTags map[string]string) error {
	key := apt.GenerateServiceTagKey(domainProject, serviceId)
	data, err := json.Marshal(dataTags)
	if err != nil {
		util.Logger().Errorf(err, "add tag into etcd,serviceId %s:json marshal tag data failed.", serviceId)
		return err
	}

	_, err = backend.Registry().Do(ctx,
		registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue(data))
	if err != nil {
		util.Logger().Errorf(err, "add tag into etcd,serviceId %s: commit tag data into etcd failed.", serviceId)
		return err
	}
	return nil
}

func GetTagsUtils(ctx context.Context, domainProject, serviceId string) (tags map[string]string, err error) {
	key := apt.GenerateServiceTagKey(domainProject, serviceId)
	opts := append(FromContext(ctx), registry.WithStrKey(key))
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
