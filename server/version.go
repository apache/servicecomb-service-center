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
package server

import (
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

func LoadServerVersion() error {
	resp, err := backend.Registry().Do(context.Background(),
		registry.GET, registry.WithStrKey(core.GetServerInfoKey()))
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

func UpgradeServerVersion() error {
	bytes, err := json.Marshal(core.ServerInfo)
	if err != nil {
		return err
	}
	_, err = backend.Registry().Do(context.Background(),
		registry.PUT, registry.WithStrKey(core.GetServerInfoKey()), registry.WithValue(bytes))
	if err != nil {
		return err
	}
	return nil
}
