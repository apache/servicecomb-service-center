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
	"encoding/json"

	"github.com/go-chassis/cari/sync"
	"github.com/little-cui/etcdadpt"

	"github.com/apache/servicecomb-service-center/eventbase/datasource/etcd/key"
)

func GenTombstoneOpts(domain, project, resourceType, resourceID string) (etcdadpt.OpOptions, error) {
	tombstone := sync.NewTombstone(domain, project, resourceType, resourceID)
	tombstoneBytes, err := json.Marshal(tombstone)
	if err != nil {
		return etcdadpt.OpOptions{}, err
	}
	tombstoneOpPut := etcdadpt.OpPut(etcdadpt.WithStrKey(key.TombstoneKey(domain, project, tombstone.ResourceType,
		tombstone.ResourceID)), etcdadpt.WithValue(tombstoneBytes))
	return tombstoneOpPut, nil
}