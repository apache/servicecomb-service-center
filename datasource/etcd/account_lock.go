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
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type AccountLockManager struct {
	releaseAfter time.Duration
}

func (al AccountLockManager) GetLock(ctx context.Context, key string) (*datasource.AccountLock, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateAccountLockKey(key)))
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, datasource.ErrAccountLockNotExist
	}
	lock := &datasource.AccountLock{}
	err = json.Unmarshal(resp.Kvs[0].Value, lock)
	if err != nil {
		log.Error(fmt.Sprintf("key %s format invalid", key), err)
		return nil, err
	}
	return lock, nil
}

func (al AccountLockManager) ListLock(ctx context.Context) ([]*datasource.AccountLock, int64, error) {
	resp, err := client.Instance().Do(ctx, client.GET,
		client.WithStrKey(path.GenerateAccountLockKey("")), client.WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	locks := make([]*datasource.AccountLock, 0, resp.Count)
	for _, v := range resp.Kvs {
		lock := &datasource.AccountLock{}
		err = json.Unmarshal(v.Value, lock)
		if err != nil {
			log.Error("account lock info format invalid:", err)
			continue //do not fail if some account is invalid
		}
		locks = append(locks, lock)
	}
	return locks, resp.Count, nil
}

func (al AccountLockManager) DeleteLock(ctx context.Context, key string) error {
	_, err := client.Delete(ctx, path.GenerateAccountLockKey(key))
	if err != nil {
		log.Error(fmt.Sprintf("remove lock %s failed", key), err)
		return datasource.ErrCannotReleaseLock
	}
	log.Info(fmt.Sprintf("%s is released", key))
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
	value, err := json.Marshal(l)
	if err != nil {
		log.Errorf(err, "account lock is invalid")
		return err
	}
	etcdKey := path.GenerateAccountLockKey(key)
	err = client.PutBytes(ctx, etcdKey, value)
	if err != nil {
		log.Errorf(err, "can not save account lock")
		return err
	}
	log.Info(fmt.Sprintf("%s is locked, release at %d", key, l.ReleaseAt))
	return nil
}
