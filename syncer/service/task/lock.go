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

package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	serverconfig "github.com/apache/servicecomb-service-center/server/config"
	"github.com/go-chassis/cari/dlock"

	// glint
	_ "github.com/go-chassis/cari/dlock/bootstrap"
	"github.com/go-chassis/foundation/gopool"
)

type DistributedLock struct {
	heartbeatDuration time.Duration
	ttl               int64
	do                func(ctx context.Context)
	key               string
	isLock            bool

	locker dlock.DLock
}

func (dl *DistributedLock) LockDo() {
	gopool.Go(func(goctx context.Context) {
		log.Info("start lock key " + dl.key)
		ticker := time.NewTicker(dl.heartbeatDuration)
		var ctx context.Context
		var cancel context.CancelFunc
		failCount := 0
		dl.newLock()

		for {
			select {
			case <-ticker.C:
				if !dl.isLock {
					err := dl.locker.TryLock(dl.key, dl.ttl)
					if err != nil {
						continue
					}
					log.Info(fmt.Sprintf("lock key %s success", dl.key))

					ctx, cancel = context.WithCancel(context.Background())
					dl.do(ctx)
					dl.isLock = true
					continue
				}

				err := dl.locker.Renew(dl.key)
				if err == nil {
					log.Info(fmt.Sprintf("renew lock %s success", dl.key))
					continue
				}

				if !errors.Is(err, dlock.ErrDLockNotExists) {
					failCount++
					log.Error("renew lock failed", err)
					if failCount == 5 {
						log.Warn("renew lock failed 5 times, release lock")
						cancel()
						dl.isLock = false
						failCount = 0
					}
					continue
				}
			case <-goctx.Done():
				ticker.Stop()
				cancel()
				log.Info(fmt.Sprintf("release lock %s", dl.key))
				return
			}
		}
	})
}

func (dl *DistributedLock) newLock() {
	kind := serverconfig.GetString("registry.kind", "",
		serverconfig.WithStandby("registry_plugin"))

	for {
		err := dlock.Init(dlock.Options{
			Kind: kind,
		})
		if err == nil {
			dl.locker = dlock.Instance()
			log.Info("init lock success")
			break
		}

		log.Warn(fmt.Sprintf("init dlock failed, %s", err.Error()))
		time.Sleep(5 * time.Second)
	}
}
