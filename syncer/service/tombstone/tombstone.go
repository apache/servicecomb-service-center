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
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/eventbase/model"
	"github.com/apache/servicecomb-service-center/eventbase/service/tombstone"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
)

const (
	defaultExpireTime = time.Hour * 24
)

func DeleteExpireTombStone() error {
	expireTime := getExpireTime()
	request := &model.ListTombstoneRequest{
		BeforeTimestamp: time.Now().Add(-expireTime).UnixNano(),
	}
	tombstones, err := tombstone.List(context.Background(), request)
	if err != nil {
		log.Error("get tombstone list fail", err)
		return err
	}

	if len(tombstones) <= 0 {
		log.Info("expire tombstone data is empty")
		return nil
	}

	err = tombstone.Delete(context.Background(), tombstones...)
	if err != nil {
		log.Error("delete tombstone data list fail", err)
		return err
	}
	return nil
}

func getExpireTime() time.Duration {
	config := config.GetConfig()
	if config.Sync == nil || config.Sync.Tombstone == nil {
		log.Warn("tombstone expireTime is empty")
		return defaultExpireTime
	}

	exprTimeStr := config.Sync.Tombstone.ExpireTime
	if len(exprTimeStr) <= 0 {
		log.Warn("tombstone expireTime is empty")
		return defaultExpireTime
	}

	expireTime, err := time.ParseDuration(exprTimeStr)
	if err != nil {
		log.Error(fmt.Sprintf("tombstone expireTime parseDuration expire:%s", exprTimeStr), err)
		return defaultExpireTime
	}
	return expireTime
}
