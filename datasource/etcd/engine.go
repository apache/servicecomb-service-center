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

package etcd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/version"
	"github.com/go-chassis/cari/dlock"
	"github.com/little-cui/etcdadpt"
)

const versionLockKey = "/version-upgrade"

type SCManager struct {
}

func (sm *SCManager) GetClusters(ctx context.Context) (etcdadpt.Clusters, error) {
	return etcdadpt.ListCluster(ctx)
}
func (sm *SCManager) UpgradeServerVersion(ctx context.Context) error {
	bytes, err := json.Marshal(config.Server)
	if err != nil {
		return err
	}
	return etcdadpt.PutBytes(ctx, path.GetServerInfoKey(), bytes)
}
func (sm *SCManager) UpgradeVersion(ctx context.Context) error {
	if err := dlock.Lock(versionLockKey, -1); err != nil {
		log.Error("wait for server ready failed", err)
		return err
	}
	defer func() {
		if err := dlock.Unlock(versionLockKey); err != nil {
			log.Error("unlock failed", err)
		}
	}()

	if needUpgrade(ctx) {
		config.Server.Version = version.Ver().Version
		if err := sm.UpgradeServerVersion(ctx); err != nil {
			log.Error("upgrade server version failed", err)
			os.Exit(1)
		}
	}
	return nil
}
