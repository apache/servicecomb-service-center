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
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/infra/registry"
	"github.com/ServiceComb/service-center/server/plugin"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var (
	registryInstance registry.Registry
	singletonLock    sync.Mutex
	wait_delay       = []int{1, 1, 5, 10, 20, 30, 60}
)

const (
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
				util.Logger().Errorf(err, "get register center client failed")
			}
			registryInstance = inst

			if registryInstance != nil {
				singletonLock.Unlock()
				return registryInstance
			}

			if i >= len(wait_delay) {
				i = len(wait_delay) - 1
			}
			t := time.Duration(wait_delay[i]) * time.Second
			util.Logger().Errorf(nil, "initialize service center failed, retry after %s", t)
			<-time.After(t)
		}
		singletonLock.Unlock()
	}
	return registryInstance
}

func BatchCommit(ctx context.Context, opts []registry.PluginOp) error {
	lenOpts := len(opts)
	tmpLen := lenOpts
	tmpOpts := []registry.PluginOp{}
	var err error
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MAX_TXN_NUMBER_ONE_TIME
		if tmpLen > 0 {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : (i+1)*MAX_TXN_NUMBER_ONE_TIME]
		} else {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : lenOpts]
		}
		_, err = Registry().Txn(ctx, tmpOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
