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

package heartbeatcache

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	dmongo "github.com/go-chassis/cari/db/mongo"
	"github.com/go-chassis/foundation/gopool"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	DefaultTTL              = 30
	defaultCacheCapacity    = 10000
	defaultWorkNum          = 10
	defaultTimeout          = 10
	instanceCheckerInternal = 1 * time.Second
	ctxTimeout              = 5 * time.Second
)

var ErrHeartbeatTimeout = errors.New("heartbeat task waiting for processing timeout. ")

var (
	once sync.Once
	cfg  CacheConfig
)

type CacheConfig struct {
	CacheChan              chan *InstanceHeartbeatInfo
	InstanceHeartbeatStore *cache.Cache
	WorkerNum              int
	HeartbeatTaskTimeout   int
}

func Configuration() *CacheConfig {
	once.Do(func() {
		cfg.WorkerNum = runtime.NumCPU()
		num := config.GetInt("heartbeat.workerNum", defaultWorkNum)
		if num != 0 {
			cfg.WorkerNum = num
		}
		cfg.HeartbeatTaskTimeout = config.GetInt("heartbeat.timeout", defaultTimeout)
		cfg.CacheChan = make(chan *InstanceHeartbeatInfo, config.GetInt("heartbeat.cacheCapacity", defaultCacheCapacity))
		cfg.InstanceHeartbeatStore = cache.New(0, instanceCheckerInternal)
		cfg.InstanceHeartbeatStore.OnEvicted(func(k string, v interface{}) {
			instanceInfo, ok := v.(*InstanceHeartbeatInfo)
			if ok && instanceInfo != nil {
				ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
				defer cancel()
				err := cleanInstance(ctx, instanceInfo.ServiceID, instanceInfo.InstanceID)
				if err != nil {
					log.Error("failed to cleanInstance in mongodb.", err)
				}
			}
		})
		for i := 1; i <= cfg.WorkerNum; i++ {
			gopool.Go(func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						log.Warn("heartbeat work protocol exit.")
						return
					case heartbeatInfo, ok := <-cfg.CacheChan:
						if ok {
							cfg.InstanceHeartbeatStore.Set(heartbeatInfo.InstanceID, heartbeatInfo, time.Duration(heartbeatInfo.TTL)*time.Second)
						}
					}
				}
			})
		}
	})
	return &cfg
}

func (c *CacheConfig) AddHeartbeatTask(serviceID string, instanceID string, ttl int32) error {
	// Unassigned setting default value is 30s
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	newInstance := &InstanceHeartbeatInfo{
		ServiceID:   serviceID,
		InstanceID:  instanceID,
		TTL:         ttl,
		LastRefresh: time.Now(),
	}
	select {
	case c.CacheChan <- newInstance:
		return nil
	case <-time.After(time.Duration(c.HeartbeatTaskTimeout) * time.Second):
		log.Warn("the heartbeat's channel is full. ")
		return ErrHeartbeatTimeout
	}
}

func (c *CacheConfig) RemoveCacheInstance(instanceID string) {
	c.InstanceHeartbeatStore.Delete(instanceID)
}

func cleanInstance(ctx context.Context, serviceID string, instanceID string) error {
	var cleanFunc = func(sessionContext mongo.SessionContext) error {
		filter := util.NewFilter(util.InstanceServiceID(serviceID), util.InstanceInstanceID(instanceID))
		result := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).FindOne(ctx, filter)
		if result.Err() != nil {
			log.Error("failed to query instance", result.Err())
			return result.Err()
		}
		var ins model.Instance
		err := result.Decode(&ins)
		if err != nil {
			log.Error("decode instance failed", err)
			return err
		}
		ttl := ins.Instance.HealthCheck.Interval * (ins.Instance.HealthCheck.Times + 1)
		if ttl <= 0 {
			ttl = DefaultTTL
		}
		if isOutDate(ins.RefreshTime, ttl) {
			return nil
		}
		err = removeDBInstance(ctx, ins.Instance.ServiceId, ins.Instance.InstanceId)
		if err != nil {
			log.Error("fail to remote instance in db", err)
		}
		return err
	}
	return dmongo.GetClient().ExecTxn(ctx, cleanFunc)
}

func removeDBInstance(ctx context.Context, serviceID string, instanceID string) error {
	filter := util.NewFilter(util.InstanceServiceID(serviceID), util.InstanceInstanceID(instanceID))
	res, err := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).DeleteOne(ctx, filter)
	if err != nil {
		log.Error("failed to clean instance", err)
		return err
	}
	log.Info(fmt.Sprintf("delete from mongodb:%+v", res))
	return nil
}

func findInstance(ctx context.Context, serviceID string, instanceID string) (*model.Instance, error) {
	filter := util.NewFilter(util.InstanceServiceID(serviceID), util.InstanceInstanceID(instanceID))
	result := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var ins model.Instance
	err := result.Decode(&ins)
	if err != nil {
		log.Error("decode instance failed", err)
		return nil, err
	}
	return &ins, nil
}

func updateInstance(ctx context.Context, serviceID string, instanceID string) error {
	filter := util.NewFilter(util.InstanceServiceID(serviceID), util.InstanceInstanceID(instanceID))
	update := bson.M{
		"$set": bson.M{model.ColumnRefreshTime: time.Now()},
	}
	result := dmongo.GetClient().GetDB().Collection(model.CollectionInstance).FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		log.Error("failed to update refresh time of instance", result.Err())
	}
	return result.Err()
}

func isOutDate(refreshTime time.Time, ttl int32) bool {
	return time.Since(refreshTime) <= time.Duration(ttl)*time.Second
}
