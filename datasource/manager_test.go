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

package datasource_test

// initialize
import _ "github.com/apache/servicecomb-service-center/server/bootstrap"
import (
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	_ = archaius.Init(archaius.WithMemorySource())
	_ = archaius.Set("servicecomb.datasource.name", "etcd")
	_ = archaius.Set("servicecomb.instance.TTL", 1000)
	_ = archaius.Set("servicecomb.instance.editable", "true")
	t.Run("init microservice data source plugin, should not pass", func(t *testing.T) {
		schemaEditableConfig := strings.ToLower(archaius.GetString("servicecomb.schema.editable", "true"))
		schemaEditable := strings.Compare(schemaEditableConfig, "true") == 0
		pluginName := datasource.ImplName("unknown")
		TTL, err := strconv.ParseInt(archaius.GetString("servicecomb.instance.TTL", "1000"), 10, 0)
		if err != nil {
			log.Error("microservice etcd implement failed for INSTANCE_TTL config: %v", err)
		}

		err = datasource.Init(datasource.Options{
			Endpoint:       "",
			PluginImplName: pluginName,
			TTL:            TTL,
			SchemaEditable: schemaEditable,
		})
		assert.Error(t, err)
	})
	t.Run("install and init microservice data source plugin, should pass", func(t *testing.T) {
		datasource.Install("etcd", func(opts datasource.Options) (datasource.DataSource, error) {
			return etcd.NewDataSource(opts)
		})

		schemaEditableConfig := strings.ToLower(archaius.GetString("servicecomb.schema.editable", "true"))
		schemaEditable := strings.Compare(schemaEditableConfig, "true") == 0
		pluginName := datasource.ImplName(archaius.GetString("servicecomb.datasource.name", "etcd"))
		TTL, err := strconv.ParseInt(archaius.GetString("servicecomb.instance.TTL", "1000"), 10, 0)
		if err != nil {
			log.Error("microservice etcd implement failed for INSTANCE_TTL config: %v", err)
		}

		err = datasource.Init(datasource.Options{
			Endpoint:       "",
			PluginImplName: pluginName,
			TTL:            TTL,
			SchemaEditable: schemaEditable,
		})
		assert.NoError(t, err)
	})
}
