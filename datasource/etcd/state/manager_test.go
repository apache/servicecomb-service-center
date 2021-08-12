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

package state_test

import (
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state"
	"github.com/little-cui/etcdadpt"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	etcdadpt.Init(etcdadpt.Config{
		Kind:             "etcd",
		ClusterAddresses: "127.0.0.1:2379",
	})
	t.Run("init microservice data source plugin, should not pass", func(t *testing.T) {
		pluginName := "unknown"
		err := state.Init(state.Config{
			Kind: pluginName,
		})
		assert.Error(t, err)
	})
	t.Run("install and init microservice data source plugin, should pass", func(t *testing.T) {
		pluginName := "etcd"
		err := state.Init(state.Config{
			Kind: pluginName,
		})
		assert.NoError(t, err)
	})
}
