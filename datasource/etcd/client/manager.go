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

package client

import (
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/backoff"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type newClientFunc func(opts datasource.Options) Registry

var (
	plugins    = make(map[datasource.Kind]newClientFunc)
	pluginInst Registry
)

// load plugins configuration into plugins
func Install(pluginImplName string, newFunc newClientFunc) {
	plugins[datasource.Kind(pluginImplName)] = newFunc
}

// construct storage plugin instance
// invoked by sc main process
func Init(opts datasource.Options) error {
	for i := 0; ; i++ {
		inst, err := New(opts)
		if err == nil {
			pluginInst = inst
			break
		}

		t := backoff.GetBackoff().Delay(i)
		log.Errorf(err, "initialize client[%v] failed, retry after %s", opts.Kind, t)
		<-time.After(t)
	}
	log.Info(fmt.Sprintf("client plugin [%s] enabled", opts.Kind))
	return nil
}

func New(opts datasource.Options) (Registry, error) {
	if opts.Kind == "" {
		return nil, fmt.Errorf("plugin implement name is nil")
	}

	f, ok := plugins[opts.Kind]
	if !ok {
		return nil, fmt.Errorf("plugin implement not supported [%s]", opts.Kind)
	}
	inst := f(opts)
	select {
	case err := <-inst.Err():
		return nil, err
	case <-inst.Ready():
		return inst, nil
	}
}

// Instance is the instance of Etcd client
func Instance() Registry {
	if pluginInst != nil {
		<-pluginInst.Ready()
	}
	return pluginInst
}
