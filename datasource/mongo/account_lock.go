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

package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type AccountLockManager struct {
	releaseAfter time.Duration
	locks        sync.Map
}

func (al *AccountLockManager) GetLock(ctx context.Context, key string) (*datasource.AccountLock, error) {
	l, ok := al.locks.Load(key)
	if !ok {
		log.Debug(fmt.Sprintf("%s is not locked", key))
		return nil, datasource.ErrAccountLockNotExist
	}
	return l.(*datasource.AccountLock), nil
}

func (al *AccountLockManager) DeleteLock(ctx context.Context, key string) error {
	al.locks.Delete(key)
	log.Warn(fmt.Sprintf("%s is released", key))
	return nil
}

func NewAccountLockManager(ReleaseAfter time.Duration) datasource.AccountLockManager {
	return &AccountLockManager{releaseAfter: ReleaseAfter}
}

func (al *AccountLockManager) Ban(ctx context.Context, key string) error {
	l := &datasource.AccountLock{}
	l.Key = key
	l.Status = datasource.StatusBanned
	l.ReleaseAt = time.Now().Add(al.releaseAfter).Unix()
	al.locks.Store(key, l)
	log.Warn(fmt.Sprintf("%s is locked, release at %d", key, l.ReleaseAt))
	return nil
}
