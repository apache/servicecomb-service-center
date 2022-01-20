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
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/syncer/service/tombstone"
	"github.com/go-chassis/cari/dlock"
	"github.com/robfig/cron/v3"
)

const (
	defaultDeleteExpireTombstoneCron = "0 0 * * *" // once every day at 00:00
	deleteExpireTombstoneLockTTL     = 60
	deleteExpireTombstoneLockKey     = "delete-expireTombstone-job"
)

func init() {
	cronStr := config.GetString("sync.tombstone.retire,cron", "")
	if len(cronStr) <= 0 {
		cronStr = defaultDeleteExpireTombstoneCron
	}

	log.Info(fmt.Sprintf("start syncer tombstone job, plan is %v", cronStr))
	c := cron.New()
	_, err := c.AddFunc(cronStr, func() {
		deleteExpireTombStone()
	})
	if err != nil {
		log.Error("cron add func failed", err)
		return
	}
	c.Start()
}

func deleteExpireTombStone() {
	err := dlock.TryLock(deleteExpireTombstoneLockKey, deleteExpireTombstoneLockTTL)
	if err != nil {
		log.Error(fmt.Sprintf("try lock %s failed", deleteExpireTombstoneLockKey), err)
		return
	}
	defer func() {
		if err := dlock.Unlock(deleteExpireTombstoneLockKey); err != nil {
			log.Error("unlock failed", err)
		}
	}()

	log.Info("start delete expire tombstone job")
	err = tombstone.DeleteExpireTombStone()
	if err != nil {
		log.Error("delete expire tombstone failed", err)
	}
}
