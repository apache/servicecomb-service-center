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

package tombstone_test

import (
	"context"
	"testing"

	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	_ "github.com/apache/servicecomb-service-center/eventbase/test"
)

func TestTombstoneService(t *testing.T) {
	tombstoneOne := sync.NewTombstone("tombstone", "tombstone", "config", "111111")
	tombstoneTwo := sync.NewTombstone("tombstone", "tombstone", "config", "222222")
	tombstoneThree := sync.NewTombstone("tombstone", "tombstone", "config", "333333")

	t.Run("to create three tasks for next get delete and list operations, should pass", func(t *testing.T) {
		_, err := datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneOne)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneTwo)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneThree)
		assert.Nil(t, err)
	})

	t.Run("get tombstone service", func(t *testing.T) {
		t.Run("get tombstoneOne should pass", func(t *testing.T) {
			getReq := model.GetTombstoneRequest{
				Project:      tombstoneOne.Project,
				Domain:       tombstoneOne.Domain,
				ResourceType: tombstoneOne.ResourceType,
				ResourceID:   tombstoneOne.ResourceID,
			}
			tmpTombstone, err := tombstone.Get(context.Background(), &getReq)
			assert.Nil(t, err)
			assert.Equal(t, tmpTombstone.ResourceID, tmpTombstone.ResourceID)
			assert.Equal(t, tmpTombstone.ResourceType, tmpTombstone.ResourceType)
			assert.Equal(t, tmpTombstone.Domain, tmpTombstone.Domain)
			assert.Equal(t, tmpTombstone.Project, tmpTombstone.Project)
		})
	})

	t.Run("list tombstone service", func(t *testing.T) {
		t.Run("list all tombstones in default domain and default project should pass", func(t *testing.T) {
			listReq := model.ListTombstoneRequest{
				Domain:  "tombstone",
				Project: "tombstone",
			}
			tombstones, err := tombstone.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 3, len(tombstones))
		})
	})

	t.Run("delete tombstone service", func(t *testing.T) {
		t.Run("delete all tombstones in default domain and default project should pass", func(t *testing.T) {
			listReq := model.ListTombstoneRequest{
				Domain:  "tombstone",
				Project: "tombstone",
			}
			tombstones, err := tombstone.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 3, len(tombstones))
			err = tombstone.Delete(context.Background(), tombstones...)
			assert.Nil(t, err)
			dTombstones, err := tombstone.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 0, len(dTombstones))
		})
	})

}
