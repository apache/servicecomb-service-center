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
	"encoding/json"

	"github.com/go-chassis/cari/sync"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/datasource/etcd/key"
	"github.com/apache/servicecomb-service-center/eventbase/model"
)

type Dao struct {
}

func (d *Dao) Get(ctx context.Context, req *model.GetTombstoneRequest) (*sync.Tombstone, error) {
	tombstoneKey := key.TombstoneKey(req.Domain, req.Project, req.ResourceType, req.ResourceID)
	kv, err := etcdadpt.Get(ctx, tombstoneKey)
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, datasource.ErrTombstoneNotExists
	}
	tombstone := sync.Tombstone{}
	err = json.Unmarshal(kv.Value, &tombstone)
	if err != nil {
		return nil, err
	}
	return &tombstone, nil
}

func (d *Dao) Create(ctx context.Context, tombstone *sync.Tombstone) (*sync.Tombstone, error) {
	tombstoneBytes, err := json.Marshal(tombstone)
	if err != nil {
		return nil, err
	}
	ok, err := etcdadpt.InsertBytes(ctx, key.TombstoneKey(tombstone.Domain, tombstone.Project,
		tombstone.ResourceType, tombstone.ResourceID), tombstoneBytes)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, datasource.ErrTombstoneAlreadyExists
	}
	return tombstone, nil
}

func (d *Dao) Delete(ctx context.Context, tombstones ...*sync.Tombstone) error {
	delOptions := make([]etcdadpt.OpOptions, len(tombstones))
	for i, tombstone := range tombstones {
		delOptions[i] = etcdadpt.OpDel(etcdadpt.WithStrKey(key.TombstoneKey(tombstone.Domain, tombstone.Project,
			tombstone.ResourceType, tombstone.ResourceID)))
	}
	err := etcdadpt.Txn(ctx, delOptions)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) List(ctx context.Context, options ...datasource.TombstoneFindOption) ([]*sync.Tombstone, error) {
	opts := datasource.NewTombstoneFindOptions()
	for _, o := range options {
		o(&opts)
	}
	tombstones := make([]*sync.Tombstone, 0)
	kvs, _, err := etcdadpt.List(ctx, key.TombstoneList(opts.Domain, opts.Project))
	if err != nil {
		return tombstones, err
	}
	for _, kv := range kvs {
		tombstone := sync.Tombstone{}
		err = json.Unmarshal(kv.Value, &tombstone)
		if err != nil {
			continue
		}
		if !filterMatch(&tombstone, opts) {
			continue
		}
		formatTombstoneKey(&tombstone, kv.Key)
		tombstones = append(tombstones, &tombstone)
	}
	return tombstones, nil
}

func filterMatch(tombstone *sync.Tombstone, options datasource.TombstoneFindOptions) bool {
	if options.ResourceType != "" && tombstone.ResourceType != options.ResourceType {
		return false
	}
	if options.BeforeTimestamp != 0 && tombstone.Timestamp > options.BeforeTimestamp {
		return false
	}
	return true
}

func formatTombstoneKey(tombstone *sync.Tombstone, k []byte) {
	keyInfo := key.SplitTombstoneKey(k)
	if len(keyInfo) <= key.TombstoneKeyLen {
		return
	}
	if len(tombstone.Domain) <= 0 {
		tombstone.Domain = keyInfo[3]
	}
	if len(tombstone.Project) <= 0 {
		tombstone.Project = keyInfo[4]
	}
	if len(tombstone.ResourceType) <= 0 {
		tombstone.ResourceType = keyInfo[5]
	}
	if len(tombstone.ResourceID) <= 0 {
		tombstone.ResourceID = key.JoinResourceID(keyInfo[key.TombstoneKeyLen:])
	}
	return
}
