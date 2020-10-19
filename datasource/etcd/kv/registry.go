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

package kv

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/datasource"
	registry "github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"sync"
	"time"
)

var (
	engineInstance *RegistryEngine
	singletonLock  sync.Mutex
)

const (
	// the same as v3rpc.MaxOpsPerTxn = 128
	MaxTxnNumberOneTime = 128
)

func NewEngine() (*RegistryEngine, error) {
	instance := plugin.Plugins().Registry()
	if instance == nil {
		return nil, errors.New("register center client plugin does not exist")
	}
	select {
	case err := <-instance.Err():
		plugin.Plugins().Reload(plugin.REGISTRY)
		return nil, err
	case <-instance.Ready():
	}
	return &RegistryEngine{
		Registry:  instance,
		goroutine: gopool.New(context.Background()),
	}, nil
}

func Registry() registry.Registry {
	return GetRegistryEngine()
}

func GetRegistryEngine() *RegistryEngine {
	if engineInstance == nil {
		singletonLock.Lock()
		for i := 0; engineInstance == nil; i++ {
			inst, err := NewEngine()
			if err != nil {
				log.Errorf(err, "get register center client failed")
			}
			engineInstance = inst

			if engineInstance != nil {
				singletonLock.Unlock()
				return engineInstance
			}

			t := backoff.GetBackoff().Delay(i)
			log.Errorf(nil, "initialize service center failed, retry after %s", t)
			<-time.After(t)
		}
		singletonLock.Unlock()
	}
	return engineInstance
}

func BatchCommit(ctx context.Context, opts []registry.PluginOp) error {
	_, err := BatchCommitWithCmp(ctx, opts, nil, nil)
	return err
}

func BatchCommitWithCmp(ctx context.Context, opts []registry.PluginOp,
	cmp []registry.CompareOp, fail []registry.PluginOp) (resp *registry.PluginResponse, err error) {
	lenOpts := len(opts)
	tmpLen := lenOpts
	var tmpOpts []registry.PluginOp
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MaxTxnNumberOneTime
		if tmpLen > 0 {
			tmpOpts = opts[i*MaxTxnNumberOneTime : (i+1)*MaxTxnNumberOneTime]
		} else {
			tmpOpts = opts[i*MaxTxnNumberOneTime : lenOpts]
		}
		resp, err = Registry().TxnWithCmp(ctx, tmpOpts, cmp, fail)
		if err != nil || !resp.Succeeded {
			return
		}
	}
	return
}

type RegistryEngine struct {
	registry.Registry
	goroutine *gopool.Pool
}

func (s *RegistryEngine) Start() error {
	err := datasource.Instance().SelfRegister(context.Background())
	if err != nil {
		return err
	}

	s.heartBeatService()
	ReportScInstance()
	return nil
}

func (s *RegistryEngine) Stop() {
	s.goroutine.Close(true)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := datasource.Instance().SelfUnregister(ctx); err != nil {
		log.Error("stop registry engine failed", err)
	}
}

func (s *RegistryEngine) heartBeatService() {
	s.goroutine.Do(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				_ = datasource.Instance().SelfHeartBeat(ctx)
			}
		}
	})
}
