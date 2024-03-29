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

	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/version"
)

func loadServerVersion(ctx context.Context) error {
	kv, err := etcdadpt.Get(ctx, path.GetServerInfoKey())
	if err != nil {
		return err
	}
	if kv == nil {
		return nil
	}

	err = json.Unmarshal(kv.Value, &config.Server)
	if err != nil {
		log.Error("load server version failed, maybe incompatible", err)
		return nil
	}
	return nil
}

func needUpgrade(ctx context.Context) bool {
	err := loadServerVersion(ctx)
	if err != nil {
		log.Error("check version failed, can not load the system config", err)
		return false
	}

	update := !serviceUtil.VersionMatchRule(config.Server.Version,
		fmt.Sprintf("%s+", version.Ver().Version))
	if !update && version.Ver().Version != config.Server.Version {
		log.Warn(fmt.Sprintf("there is a higher version '%s' in cluster, now running '%s' version may be incompatible",
			config.Server.Version, version.Ver().Version))
	}

	return update
}
