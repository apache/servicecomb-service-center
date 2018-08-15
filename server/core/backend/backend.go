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
package backend

import (
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var (
	registryInstance registry.Registry
	singletonLock    sync.Mutex
)

const (
	// the same as v3rpc.MaxOpsPerTxn = 128
	MAX_TXN_NUMBER_ONE_TIME = 128
)

func New() (registry.Registry, error) {
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
	return instance, nil
}

func Registry() registry.Registry {
	if registryInstance == nil {
		singletonLock.Lock()
		for i := 0; registryInstance == nil; i++ {
			inst, err := New()
			if err != nil {
				log.Errorf(err, "get register center client failed")
			}
			registryInstance = inst

			if registryInstance != nil {
				singletonLock.Unlock()
				return registryInstance
			}

			t := util.GetBackoff().Delay(i)
			log.Errorf(nil, "initialize service center failed, retry after %s", t)
			<-time.After(t)
		}
		singletonLock.Unlock()
	}
	return registryInstance
}

func BatchCommit(ctx context.Context, opts []registry.PluginOp) error {
	_, err := BatchCommitWithCmp(ctx, opts, nil, nil)
	return err
}

func BatchCommitWithCmp(ctx context.Context, opts []registry.PluginOp,
	cmp []registry.CompareOp, fail []registry.PluginOp) (resp *registry.PluginResponse, err error) {
	lenOpts := len(opts)
	tmpLen := lenOpts
	tmpOpts := []registry.PluginOp{}
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MAX_TXN_NUMBER_ONE_TIME
		if tmpLen > 0 {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : (i+1)*MAX_TXN_NUMBER_ONE_TIME]
		} else {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : lenOpts]
		}
		resp, err = Registry().TxnWithCmp(ctx, tmpOpts, cmp, fail)
		if err != nil || !resp.Succeeded {
			return
		}
	}
	return
}
