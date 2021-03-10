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

package sd

import (
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type newCacheFunc func(opts Options) AdaptorRepository

var (
	plugins    = make(map[Kind]newCacheFunc)
	pluginInst AdaptorRepository
)

// load plugins configuration into plugins
func Install(pluginImplName string, newFunc newCacheFunc) {
	plugins[Kind(pluginImplName)] = newFunc
}

// construct storage plugin instance
// invoked by sc main process
func Init(opts Options) error {
	inst, err := New(opts)
	if err != nil {
		return err
	}
	pluginInst = inst
	log.Info(fmt.Sprintf("cache plugin [%s] enabled", opts.Kind))
	return nil
}

func New(opts Options) (AdaptorRepository, error) {
	if opts.Kind == "" {
		return nil, fmt.Errorf("plugin implement name is nil")
	}

	f, ok := plugins[opts.Kind]
	if !ok {
		return nil, fmt.Errorf("plugin implement not supported [%s]", opts.Kind)
	}
	inst := f(opts)
	return inst, nil
}

// Instance is the instance of Cache
func Instance() AdaptorRepository {
	return pluginInst
}
