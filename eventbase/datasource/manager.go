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
	"time"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/openlog"
)

const (
	DefaultTimeout = 60 * time.Second
	DefaultDBKind  = "embedded_etcd"
)

var (
	dataSourceInst DataSource
	plugins        = make(map[string]dataSourceEngine)
)

type dataSourceEngine func(c *db.Config) (DataSource, error)

func GetDataSource() DataSource {
	return dataSourceInst
}

func RegisterPlugin(name string, engineFunc dataSourceEngine) {
	plugins[name] = engineFunc
}

func Init(c db.Config) error {
	var err error
	if c.Kind == "" {
		c.Kind = DefaultDBKind
	}
	f, ok := plugins[c.Kind]
	if !ok {
		return fmt.Errorf("do not support %s", c.Kind)
	}
	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}
	dbc := &db.Config{
		URI:         c.URI,
		PoolSize:    c.PoolSize,
		SSLEnabled:  c.SSLEnabled,
		RootCA:      c.RootCA,
		CertFile:    c.CertFile,
		CertPwdFile: c.CertPwdFile,
		KeyFile:     c.KeyFile,
		Timeout:     c.Timeout,
	}

	if dataSourceInst, err = f(dbc); err != nil {
		return err
	}
	openlog.Info(fmt.Sprintf("use %s as storage", c.Kind))
	return nil
}

func GetTaskDao() TaskDao {
	return dataSourceInst.TaskDao()
}

func GetTombstoneDao() TombstoneDao {
	return dataSourceInst.TombstoneDao()
}
