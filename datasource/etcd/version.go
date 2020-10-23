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
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"github.com/apache/servicecomb-service-center/version"
	"os"
)

func (ds *DataSource) LoadServerVersion(ctx context.Context) error {
	resp, err := client.Instance().Do(ctx,
		client.GET, client.WithStrKey(core.GetServerInfoKey()))
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &core.ServerInfo)
	if err != nil {
		log.Errorf(err, "load server version failed, maybe incompatible")
		return nil
	}
	return nil
}

func (ds *DataSource) UpgradeServerVersion(ctx context.Context) error {
	bytes, err := json.Marshal(core.ServerInfo)
	if err != nil {
		return err
	}
	_, err = client.Instance().Do(ctx,
		client.PUT, client.WithStrKey(core.GetServerInfoKey()), client.WithValue(bytes))
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataSource) UpgradeVersion(ctx context.Context) error {
	lock, err := mux.Lock(mux.GlobalLock)

	if err != nil {
		log.Errorf(err, "wait for server ready failed")
		return err
	}
	if ds.needUpgrade(ctx) {
		core.ServerInfo.Version = version.Ver().Version

		if err := ds.UpgradeServerVersion(ctx); err != nil {
			log.Errorf(err, "upgrade server version failed")
			os.Exit(1)
		}
	}
	err = lock.Unlock()
	if err != nil {
		log.Error("", err)
	}
	return err
}

func (ds *DataSource) needUpgrade(ctx context.Context) bool {
	err := ds.LoadServerVersion(ctx)
	if err != nil {
		log.Errorf(err, "check version failed, can not load the system config")
		return false
	}

	update := !serviceUtil.VersionMatchRule(core.ServerInfo.Version,
		fmt.Sprintf("%s+", version.Ver().Version))
	if !update && version.Ver().Version != core.ServerInfo.Version {
		log.Warnf(
			"there is a higher version '%s' in cluster, now running '%s' version may be incompatible",
			core.ServerInfo.Version, version.Ver().Version)
	}

	return update
}
