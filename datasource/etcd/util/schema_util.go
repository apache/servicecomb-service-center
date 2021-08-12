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

package util

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/little-cui/etcdadpt"
)

func CheckSchemaInfoExist(ctx context.Context, key string) (bool, error) {
	opts := append(FromContext(ctx), etcdadpt.WithStrKey(key), etcdadpt.WithCountOnly())
	resp, errDo := sd.Schema().Search(ctx, opts...)
	if errDo != nil {
		return false, errDo
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}
