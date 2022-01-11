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

package tombstone

import (
	"context"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/eventbase/test"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := datasource.Init(test.DbCfg)
	if err != nil {
		panic(err)
	}
}

const (
	testDomain  = "expireTombstone"
	testProject = "expireProject"
)

func TestDeleteExpireTombStone(t *testing.T) {
	tombstoneOne := sync.NewTombstone(testDomain, testProject, "config", "1")
	tombstoneOne.Timestamp = time.Now().Add(-time.Hour * 24).UnixNano()

	tombstoneTwo := sync.NewTombstone(testDomain, testProject, "config", "2")
	tombstoneTwo.Timestamp = time.Now().Add(-time.Hour * 23).UnixNano()

	tombstoneThree := sync.NewTombstone(testDomain, testProject, "config", "3")
	tombstoneThree.Timestamp = time.Now().Add(-time.Hour * 25).UnixNano()

	t.Run("to create three tasks for next get delete and list operations, should pass", func(t *testing.T) {
		_, err := datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneOne)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneTwo)
		assert.Nil(t, err)
		_, err = datasource.GetDataSource().TombstoneDao().Create(context.Background(), tombstoneThree)
		assert.Nil(t, err)
	})

	t.Run("list tombstone service", func(t *testing.T) {
		t.Run("list all tombstones in default domain and default project should pass", func(t *testing.T) {
			listReq := model.ListTombstoneRequest{
				Domain:          testDomain,
				Project:         testProject,
				BeforeTimestamp: time.Now().Add(-time.Hour * 24).UnixNano(),
			}
			tombstones, err := tombstone.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 2, len(tombstones))
		})
	})

	t.Run("delete expire tombstone service", func(t *testing.T) {
		t.Run("delete expire tombstones in default domain and default project should pass", func(t *testing.T) {
			err := DeleteExpireTombStone()
			assert.Nil(t, err)

			listReq := model.ListTombstoneRequest{
				Domain:  testDomain,
				Project: testProject,
			}
			tombstones, err := tombstone.List(context.Background(), &listReq)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(tombstones))
		})
	})

}
