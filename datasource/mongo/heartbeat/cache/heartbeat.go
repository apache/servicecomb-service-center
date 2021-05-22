/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except request compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to request writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package heartbeatcache

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	defaultTTL              = 30
	defaultCacheCapacity    = 10000
	defaultWorkNum          = 10
	defaultTimeout          = 10
	instanceCheckerInternal = 1 * time.Second
	ctxTimeout              = 5 * time.Second
)

var ErrHeartbeatTimeout = errors.New("heartbeat task waiting for processing timeout. ")

var (
	once sync.Once
	cfg  cacheConfig
)

type cacheConfig struct {
	cacheChan              chan *instanceHeartbeatInfo
	instanceHeartbeatStore *cache.Cache
	workerNum              int
	heartbeatTaskTimeout   int
}

func configuration() *cacheConfig {
	once.Do(func() {
		cfg.workerNum = runtime.NumCPU()
		num := config.GetInt("heartbeat.workerNum", defaultWorkNum)
		if num != 0 {
			cfg.workerNum = num
		}
		cfg.heartbeatTaskTimeout = config.GetInt("heartbeat.timeout", defaultTimeout)
		cfg.cacheChan = make(chan *instanceHeartbeatInfo, config.GetInt("heartbeat.cacheCapacity", defaultCacheCapacity))
		cfg.instanceHeartbeatStore = cache.New(0, instanceCheckerInternal)
		cfg.instanceHeartbeatStore.OnEvicted(func(k string, v interface{}) {
			instanceInfo, ok := v.(*instanceHeartbeatInfo)
			if ok && instanceInfo != nil {
				ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
				defer cancel()
				err := cleanInstance(ctx, instanceInfo.serviceID, instanceInfo.instanceID)
				if err != nil {
					log.Error("failed to cleanInstance in mongodb.", err)
				}
			}
		})
		for i := 1; i <= cfg.workerNum; i++ {
			gopool.Go(func(ctx context.Context) {
				for {
					select {
					case <-ctx.Done():
						log.Warn("heartbeat work protocol exit.")
						return
					case heartbeatInfo, ok := <-cfg.cacheChan:
						if ok {
							cfg.instanceHeartbeatStore.Set(heartbeatInfo.instanceID, heartbeatInfo, time.Duration(heartbeatInfo.ttl)*time.Second)
						}
					}
				}
			})
		}
	})
	return &cfg
}

func (c *cacheConfig) AddHeartbeatTask(serviceID string, instanceID string, ttl int32) error {
	// Unassigned setting default value is 30s
	if ttl <= 0 {
		ttl = defaultTTL
	}
	newInstance := &instanceHeartbeatInfo{
		serviceID:   serviceID,
		instanceID:  instanceID,
		ttl:         ttl,
		lastRefresh: time.Now(),
	}
	select {
	case c.cacheChan <- newInstance:
		return nil
	case <-time.After(time.Duration(c.heartbeatTaskTimeout) * time.Second):
		log.Warn("the heartbeat's channel is full. ")
		return ErrHeartbeatTimeout
	}
}

func (c *cacheConfig) RemoveCacheInstance(instanceID string) {
	c.instanceHeartbeatStore.Delete(instanceID)
}

func cleanInstance(ctx context.Context, serviceID string, instanceID string) error {
	session, err := client.GetMongoClient().StartSession(ctx)
	if err != nil {
		return err
	}
	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer session.EndSession(ctx)

	filter := heartbeat.NewServiceIDInstanceIDFilter(serviceID, instanceID)
	ins, err := dao.GetInstance(ctx, filter)
	if err != nil {
		log.Error("failed to query instance", err)
		return err
	}
	ttl := ins.Instance.HealthCheck.Interval * (ins.Instance.HealthCheck.Times + 1)
	if ttl <= 0 {
		ttl = defaultTTL
	}
	if isOutDate(ins.RefreshTime, ttl) {
		return nil
	}
	err = removeDBInstance(ctx, ins.Instance.ServiceId, ins.Instance.InstanceId)
	if err != nil {
		log.Error("fail to remote instance in db", err)
		errAbort := session.AbortTransaction(ctx)
		if errAbort != nil {
			return errAbort
		}
		return err
	}
	err = session.CommitTransaction(ctx)
	return err
}

func removeDBInstance(ctx context.Context, serviceID string, instanceID string) error {
	filter := heartbeat.NewServiceIDInstanceIDFilter(serviceID, instanceID)
	return dao.DeleteInstance(ctx, filter)
}

func findInstance(ctx context.Context, serviceID string, instanceID string) (*dao.Instance, error) {
	filter := heartbeat.NewServiceIDInstanceIDFilter(serviceID, instanceID)
	return dao.GetInstance(ctx, filter)
}

func updateInstance(ctx context.Context, serviceID string, instanceID string) error {
	filter := heartbeat.NewServiceIDInstanceIDFilter(serviceID, instanceID)
	setValue := bson.D{
		{dao.ColumnRefreshTime, time.Now()},
	}
	updateFilter := bson.D{
		{"$set", setValue},
	}
	result, err := client.GetMongoClient().FindOneAndUpdate(ctx, dao.CollectionInstance, filter, updateFilter)
	if err != nil {
		log.Error("failed to update refresh time of instance", err)
		return err
	}
	return result.Err()
}

func isOutDate(refreshTime time.Time, ttl int32) bool {
	return time.Since(refreshTime) <= time.Duration(ttl)*time.Second
}
