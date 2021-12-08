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

package etcd

import (
	"crypto/tls"
	"fmt"

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
	_ "github.com/little-cui/etcdadpt/embedded"
	_ "github.com/little-cui/etcdadpt/remote"

	"servicecomb-service-center/eventbase/datasource"
	"servicecomb-service-center/eventbase/datasource/etcd/task"
	"servicecomb-service-center/eventbase/datasource/etcd/tombstone"
	"servicecomb-service-center/eventbase/datasource/tlsutil"
)

type Datasource struct {
	taskDao   datasource.TaskDao
	tombstone datasource.TombstoneDao
}

func (d *Datasource) TaskDao() datasource.TaskDao {
	return d.taskDao
}

func (d *Datasource) TombstoneDao() datasource.TombstoneDao {
	return d.tombstone
}

func NewDatasource(c *db.Config) (datasource.DataSource, error) {
	openlog.Info(fmt.Sprintf("use %s as storage", c.Kind))
	var tlsConfig *tls.Config
	if c.SSLEnabled {
		var err error
		tlsConfig, err = tlsutil.Config(c)
		if err != nil {
			return nil, err
		}
	}
	inst := &Datasource{}
	inst.taskDao = &task.Dao{}
	inst.tombstone = &tombstone.Dao{}
	return inst, etcdadpt.Init(etcdadpt.Config{
		Kind:             c.Kind,
		ClusterAddresses: c.URI,
		SslEnabled:       c.SSLEnabled,
		TLSConfig:        tlsConfig,
	})
}

func init() {
	datasource.RegisterPlugin("etcd", NewDatasource)
	datasource.RegisterPlugin("embedded_etcd", NewDatasource)
}
