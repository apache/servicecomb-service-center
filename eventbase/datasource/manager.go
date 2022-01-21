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

package datasource

import (
	"fmt"

	"github.com/go-chassis/openlog"
)

var (
	dataSourceInst DataSource
	plugins        = make(map[string]dataSourceEngine)

	logger openlog.Logger
)

func Logger() openlog.Logger {
	return logger
}

type dataSourceEngine func() DataSource

func GetDataSource() DataSource {
	return dataSourceInst
}

func RegisterPlugin(name string, engineFunc dataSourceEngine) {
	plugins[name] = engineFunc
}

type Config struct {
	Kind   string
	Logger openlog.Logger
}

func Init(c *Config) error {
	f, ok := plugins[c.Kind]
	if !ok {
		return fmt.Errorf("do not support %s", c.Kind)
	}
	if c.Logger != nil {
		logger = c.Logger
	}

	dataSourceInst = f()
	return nil
}

func GetTaskDao() TaskDao {
	return dataSourceInst.TaskDao()
}

func GetTombstoneDao() TombstoneDao {
	return dataSourceInst.TombstoneDao()
}
