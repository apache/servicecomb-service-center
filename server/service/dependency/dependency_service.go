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
package dependency

import (
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/servicecomb/service-center/common/cache"
	apt "github.com/servicecomb/service-center/server/core"
	"github.com/servicecomb/service-center/server/core/registry"
	"github.com/servicecomb/service-center/util"
	"time"
)

var consumerCache *cache.Cache

func ConsumerCache() *cache.Cache {
	return consumerCache
}

/*
缓存2分钟过期
1分钟周期缓存consumers 遍历所有serviceid并查询consumers 做缓存
当发现新查询到的consumers列表变成0时则不做cache set操作
这样当consumers关系完全被删除也有1分钟的时间窗让实例变化推送到相应的consumers里 1分鐘后緩存也會自動清理
实例推送中的依赖发现实时性为T+1分钟
*/
func init() {
	d, _ := time.ParseDuration("2m")
	consumerCache = cache.New(d, d)
	go autoSyncConsumers()
}

//TODO
func autoSyncConsumers() {
	//ticker := time.NewTicker(time.Minute * 1)
	//for t := range ticker.C {
	//	util.LOGGER.Debug(fmt.Sprintf("sync consumers at %s", t))
	//	keys := microservice.GetService(context.TODO())
	//	for _, v := range keys {
	//		util.LOGGER.Debug(fmt.Sprintf("sync consumers for %s", v))
	//		domainAndId := strings.Split(v, ":::")
	//		// 查询所有consumer
	//		key := apt.GenerateProviderDependencyKey(domainAndId[0], domainAndId[1], "")
	//		resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
	//			Action:     registry.GET,
	//			Key:        []byte(key),
	//			WithPrefix: true,
	//			KeyOnly:    true,
	//		})
	//		if err != nil {
	//			util.LOGGER.Errorf(err, "query service consumers failed, provider id %s", domainAndId[1])
	//		}
	//		if len(resp.Kvs) != 0 {
	//			consumerCache.Set(v, resp.Kvs, 0)
	//		}
	//
	//	}
	//}
}
func GetConsumersInCache(tenant string, providerId string) ([]*mvccpb.KeyValue, error) {
	// 查询所有consumer
	key := apt.GenerateProviderDependencyKey(tenant, providerId, "")
	resp, err := registry.GetRegisterCenter().Do(context.TODO(), &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(key),
		WithPrefix: true,
		KeyOnly:    true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "query service consumers failed, provider id %s", providerId)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		// 如果没有，使用最后一次更新的缓存
		util.LOGGER.Debugf("query service consumers in cache %s", providerId)
		consumerList, found := consumerCache.Get(providerId)
		if found && len(consumerList.([]*mvccpb.KeyValue)) > 0 {
			return consumerList.([]*mvccpb.KeyValue), nil
		}
		return []*mvccpb.KeyValue{}, nil
	}
	consumerCache.Set(providerId, resp.Kvs, 0)

	util.LOGGER.Debugf("query service consumers in center %s", key)
	return resp.Kvs, nil
}
