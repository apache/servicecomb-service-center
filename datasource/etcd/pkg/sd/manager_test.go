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

package sd_test

// initialize
import _ "github.com/apache/servicecomb-service-center/server/bootstrap"
import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/pkg/sd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("init microservice data source plugin, should not pass", func(t *testing.T) {
		pluginName := sd.ImplName("unknown")
		err := sd.Init(sd.Options{
			PluginImplName: pluginName,
		})
		assert.Error(t, err)
	})
	t.Run("install and init microservice data source plugin, should pass", func(t *testing.T) {
		pluginName := sd.ImplName("etcd")
		err := sd.Init(sd.Options{
			PluginImplName: pluginName,
		})
		assert.NoError(t, err)
	})
}
