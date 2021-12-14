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

	"github.com/go-chassis/cari/db"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"

	"servicecomb-service-center/eventbase/datasource"
	"servicecomb-service-center/eventbase/datasource/etcd"
	"servicecomb-service-center/eventbase/model"
	"servicecomb-service-center/eventbase/test"
)

var ds datasource.DataSource

func init() {
	cfg := &db.Config{
		Kind: test.Etcd,
		URI:  test.EtcdURI,
	}
	ds, _ = etcd.NewDatasource(cfg)
}

func TestTombstone(t *testing.T) {
	var (
		tombstoneOne = sync.Tombstone{
			ResourceID:   "app/test",
			ResourceType: "config",
			Domain:       "default",
			Project:      "default",
			Timestamp:    1638171566,
		}
		tombstoneTwo = sync.Tombstone{
			ResourceID:   "property/test",
			ResourceType: "config",
			Domain:       "default",
			Project:      "default",
			Timestamp:    1638171567,
		}
	)

	t.Run("create tombstone", func(t *testing.T) {
		t.Run("create two tombstone should pass", func(t *testing.T) {
			tombstone, err := ds.TombstoneDao().Create(context.Background(), &tombstoneOne)
			assert.NoError(t, err)
			assert.NotNil(t, tombstone)
			tombstone, err = ds.TombstoneDao().Create(context.Background(), &tombstoneTwo)
			assert.NoError(t, err)
			assert.NotNil(t, tombstone)
		})
	})

	t.Run("get tombstone", func(t *testing.T) {
		t.Run("get one tombstone should pass", func(t *testing.T) {
			req := model.GetTombstoneRequest{
				Domain:       tombstoneOne.Domain,
				Project:      tombstoneOne.Project,
				ResourceType: tombstoneOne.ResourceType,
				ResourceID:   tombstoneOne.ResourceID,
			}
			tombstone, err := ds.TombstoneDao().Get(context.Background(), &req)
			assert.NoError(t, err)
			assert.Equal(t, tombstone.Timestamp, tombstoneOne.Timestamp)
		})
	})

	t.Run("list tombstone", func(t *testing.T) {
		t.Run("list tombstone with ResourceType and BeforeTimestamp should pass", func(t *testing.T) {
			opts := []datasource.TombstoneFindOption{
				datasource.WithResourceType(tombstoneOne.ResourceType),
				datasource.WithBeforeTimestamp(1638171600),
			}
			tombstones, err := ds.TombstoneDao().List(context.Background(), "default", "default", opts...)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(tombstones))
			assert.Equal(t, tombstones[0].Timestamp, tombstoneOne.Timestamp)
			assert.Equal(t, tombstones[1].Timestamp, tombstoneTwo.Timestamp)
		})
	})

	t.Run("delete tombstone", func(t *testing.T) {
		t.Run("delete two tombstone should pass", func(t *testing.T) {
			err := ds.TombstoneDao().Delete(context.Background(), []*sync.Tombstone{&tombstoneOne, &tombstoneTwo}...)
			assert.NoError(t, err)
		})
	})

}
