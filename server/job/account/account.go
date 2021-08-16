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

package account

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	accountsvc "github.com/apache/servicecomb-service-center/server/service/account"
	"github.com/go-chassis/foundation/gopool"
)

const CleanupInterval = 1 * time.Minute

func init() {
	startReleasedLockHistoryCleanupJob()
}

// startReleasedLockHistoryCleanupJob cause of accountsvc.IsBanned may be never called
// after locked account, then run a job to cleanup the released lock
func startReleasedLockHistoryCleanupJob() {
	log.Info(fmt.Sprintf("start released lock history cleanup job(every %s)", CleanupInterval))
	gopool.Go(func(ctx context.Context) {
		tick := time.NewTicker(CleanupInterval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				err := CleanupReleasedLockHistory(ctx)
				if err != nil {
					log.Error("cleanup lock history failed", err)
				}
			}
		}
	})
}

func CleanupReleasedLockHistory(ctx context.Context) error {
	locks, _, err := accountsvc.ListLock(ctx)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	var keys []string
	for _, lock := range locks {
		if lock.ReleaseAt > now {
			continue
		}
		keys = append(keys, lock.Key)
	}
	n := len(keys)
	if n == 0 {
		return nil
	}
	err = accountsvc.DeleteLockList(ctx, keys)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("cleanup %d released lock history", n))
	return nil
}
