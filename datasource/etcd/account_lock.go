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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/little-cui/etcdadpt"
)

type AccountLockManager struct {
	releaseAfter time.Duration
}

func (al AccountLockManager) UpsertLock(ctx context.Context, lock *datasource.AccountLock) error {
	value, err := json.Marshal(lock)
	if err != nil {
		log.Error("account lock is invalid", err)
		return err
	}
	key := lock.Key
	etcdKey := path.GenerateAccountLockKey(key)
	err = etcdadpt.PutBytes(ctx, etcdKey, value)
	if err != nil {
		log.Error("can not save account lock", err)
		return err
	}
	log.Info(fmt.Sprintf("%s is locked, release at %d", key, lock.ReleaseAt))
	return nil
}

func (al AccountLockManager) GetLock(ctx context.Context, key string) (*datasource.AccountLock, error) {
	kv, err := etcdadpt.Get(ctx, path.GenerateAccountLockKey(key))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, datasource.ErrAccountLockNotExist
	}
	lock := &datasource.AccountLock{}
	err = json.Unmarshal(kv.Value, lock)
	if err != nil {
		log.Error(fmt.Sprintf("key %s format invalid", key), err)
		return nil, err
	}
	return lock, nil
}

func (al AccountLockManager) ListLock(ctx context.Context) ([]*datasource.AccountLock, int64, error) {
	kvs, n, err := etcdadpt.List(ctx, path.GenerateAccountLockKey(""))
	if err != nil {
		return nil, 0, err
	}
	locks := make([]*datasource.AccountLock, 0, n)
	for _, v := range kvs {
		lock := &datasource.AccountLock{}
		err = json.Unmarshal(v.Value, lock)
		if err != nil {
			log.Error("account lock info format invalid:", err)
			continue //do not fail if some account is invalid
		}
		locks = append(locks, lock)
	}
	return locks, n, nil
}

func (al AccountLockManager) DeleteLock(ctx context.Context, key string) error {
	_, err := etcdadpt.Delete(ctx, path.GenerateAccountLockKey(key))
	if err != nil {
		log.Error(fmt.Sprintf("remove lock %s failed", key), err)
		return datasource.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%s is released", key))
	return nil
}

func (al AccountLockManager) DeleteLockList(ctx context.Context, keys []string) error {
	var opts []etcdadpt.OpOptions
	for _, key := range keys {
		opts = append(opts, etcdadpt.OpDel(etcdadpt.WithStrKey(path.GenerateAccountLockKey(key))))
	}
	if len(opts) == 0 {
		return nil
	}
	err := etcdadpt.Txn(ctx, opts)
	if err != nil {
		log.Error(fmt.Sprintf("remove locks %v failed", keys), err)
		return datasource.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%v are released", keys))
	return nil
}

func NewAccountLockManager(ReleaseAfter time.Duration) datasource.AccountLockManager {
	return &AccountLockManager{releaseAfter: ReleaseAfter}
}

func (al AccountLockManager) Ban(ctx context.Context, key string) error {
	l := &datasource.AccountLock{}
	l.Key = key
	l.Status = datasource.StatusBanned
	l.ReleaseAt = time.Now().Add(al.releaseAfter).Unix()
	return al.UpsertLock(ctx, l)
}
