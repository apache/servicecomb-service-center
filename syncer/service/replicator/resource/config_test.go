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

package resource

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	kiemodel "github.com/apache/servicecomb-kie/pkg/model"
	kiedb "github.com/apache/servicecomb-kie/server/datasource"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/go-chassis/cari/sync"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct {
	kvs map[string]*kiemodel.KVDoc
}

func (m *mockConfig) Create(ctx context.Context, doc *kiemodel.KVDoc) error {
	_, ok := m.kvs[doc.ID]
	if ok {
		return kiedb.ErrKVAlreadyExists
	}

	m.kvs[doc.ID] = doc
	return nil
}

func (m *mockConfig) Get(ctx context.Context, ID string) (*kiemodel.KVDoc, error) {
	result, ok := m.kvs[ID]
	if !ok {
		return nil, kiedb.ErrKeyNotExists
	}
	return result, nil
}

func (m *mockConfig) Update(ctx context.Context, doc *kiemodel.KVDoc) error {
	_, ok := m.kvs[doc.ID]
	if !ok {
		return kiedb.ErrKeyNotExists
	}
	m.kvs[doc.ID] = doc
	return nil
}

func (m *mockConfig) Delete(ctx context.Context, ID string) error {
	_, ok := m.kvs[ID]
	if !ok {
		return kiedb.ErrKeyNotExists
	}
	delete(m.kvs, ID)
	return nil
}

func TestOperateConfig(t *testing.T) {
	const resourceID = "test"
	t.Run("create update delete case", func(t *testing.T) {
		createTime := time.Now().Unix()
		input := &kiemodel.KVDoc{
			ID:          resourceID,
			LabelFormat: "a:b",
			Key:         "key",
			Value:       "value",
			ValueType:   "text",
			Project:     "default",
			Domain:      "default",
			Labels: map[string]string{
				"a": "b",
			},
			Status:     "enabled",
			CreateTime: createTime,
			UpdateTime: createTime,
		}
		value, _ := json.Marshal(input)
		id, _ := v1sync.NewEventID()
		e := &v1sync.Event{
			Id:        id,
			Action:    sync.CreateAction,
			Subject:   Config,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a := &kvConfig{
			event: e,
		}
		a.resource = &mockConfig{
			kvs: make(map[string]*kiemodel.KVDoc),
		}
		ctx := context.Background()
		result := a.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a.resource.Get(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
				}
			}
		}

		updateTime := time.Now().Unix()
		input = &kiemodel.KVDoc{
			ID:          resourceID,
			LabelFormat: "a:b",
			Key:         "key",
			Value:       "changed",
			ValueType:   "text",
			Project:     "default",
			Domain:      "default",
			Labels: map[string]string{
				"a": "b",
			},
			Status:     "enabled",
			CreateTime: createTime,
			UpdateTime: updateTime,
		}
		value, _ = json.Marshal(input)

		id, _ = v1sync.NewEventID()
		e1 := &v1sync.Event{
			Id:        id,
			Action:    sync.UpdateAction,
			Subject:   Config,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a1 := &kvConfig{
			event:    e1,
			resource: a.resource,
		}
		result = a1.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a1.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a1.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					data, err := a1.resource.Get(ctx, "test")
					assert.Nil(t, err)
					assert.NotNil(t, data)
					assert.Equal(t, "changed", data.Value)
				}
			}
		}

		id, _ = v1sync.NewEventID()
		e2 := &v1sync.Event{
			Id:        id,
			Action:    sync.DeleteAction,
			Subject:   Config,
			Opts:      nil,
			Value:     value,
			Timestamp: v1sync.Timestamp(),
		}
		a2 := &kvConfig{
			event:    e2,
			resource: a.resource,
		}
		result = a2.LoadCurrentResource(ctx)
		if assert.Nil(t, result) {
			result = a2.NeedOperate(ctx)
			if assert.Nil(t, result) {
				result = a2.Operate(ctx)
				if assert.NotNil(t, result) && assert.Equal(t, Success, result.Status) {
					_, err := a2.resource.Get(ctx, "test")
					assert.NotNil(t, err)
					assert.True(t, errors.Is(err, kiedb.ErrKeyNotExists))
				}
			}
		}
	})
}
