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

package schema

import (
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type engine func(opts Options) (DAO, error)

var (
	plugins  = make(map[string]engine)
	instance DAO
)

// Install load plugins configuration into plugins
func Install(pluginImplName string, engineFunc engine) {
	plugins[pluginImplName] = engineFunc
}

// Init construct storage plugin instance
// invoked by sc main process.
func Init(opts Options) error {
	if opts.Kind == "" {
		return nil
	}

	engineFunc, ok := plugins[opts.Kind]
	if !ok {
		return fmt.Errorf("plugin implement not supported [%s]", opts.Kind)
	}

	var err error
	instance, err = engineFunc(opts)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("schema plugin [%s] enabled", opts.Kind))

	return nil
}

func Instance() DAO {
	return instance
}
