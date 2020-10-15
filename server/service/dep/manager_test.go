// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dep

import (
	"github.com/apache/servicecomb-service-center/datasource"
	etcd2 "github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInit(t *testing.T) {
	_ = archaius.Init(archaius.WithMemorySource())
	_ = archaius.Set("servicecomb.dep.name", "etcd")
	t.Run("circuit dependency data source plugin", func(t *testing.T) {
		err := Init(Options{
			Endpoint:       "",
			PluginImplName: "",
		})
		assert.NoError(t, err)
	})
	t.Run(" dependency data source plugin install and init", func(t *testing.T) {
		Install("etcd",
			func(opts Options) (datasource.DataSource, error) {
				return etcd2.NewDataSource(), nil
			})

		// sc main function initialize step
		err := Init(Options{
			Endpoint:       "",
			PluginImplName: ImplName(archaius.GetString("servicecomb.dep.name", "etcd")),
		})
		assert.NoError(t, err)
	})
}
