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

package state

import (
	"fmt"
)

var (
	repoPlugins   = make(map[string]newRepoFunc)
	manager       *Manager
	configuration Config
)

type newRepoFunc func(opts Config) Repository

// Install load plugins configuration into plugins
func Install(pluginImplName string, newFunc newRepoFunc) {
	repoPlugins[pluginImplName] = newFunc
}

func Init(opts Config) error {
	opts.Init()

	configuration = opts

	inst, err := NewRepository(opts)
	if err != nil {
		return err
	}
	opts.Logger.Info(fmt.Sprintf("state plugin [%s] enabled", opts.Kind))

	manager = &Manager{
		Repository: inst,
	}
	manager.Initialize()
	manager.Run()
	<-manager.Ready()
	return nil
}

func NewRepository(opts Config) (Repository, error) {
	if opts.Kind == "" {
		return nil, fmt.Errorf("plugin implement name is nil")
	}

	f, ok := repoPlugins[opts.Kind]
	if !ok {
		return nil, fmt.Errorf("plugin implement not supported [%s]", opts.Kind)
	}
	inst := f(opts)
	return inst, nil
}

func Instance() *Manager {
	return manager
}

func Revision() int64 {
	return manager.Rev
}

func Configuration() Config {
	return configuration
}
