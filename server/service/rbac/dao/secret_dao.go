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

package dao

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/kv"
)

//OverrideSecret write secret to kv store
func OverrideSecret(ctx context.Context, sk string) error {
	key := core.GenerateRBACSecretKey()
	err := kv.Put(context.Background(), key, sk)
	if err != nil {
		log.Errorf(err, "can not override secret")
		return err
	}
	return nil
}
func GetSecret(ctx context.Context) ([]byte, error) {
	key := core.GenerateRBACSecretKey()
	r, err := kv.Get(ctx, key)
	if err != nil {
		log.Errorf(err, "can not get secret")
		return nil, err
	}
	return r.Value, nil
}
func SecretExist(ctx context.Context) (bool, error) {
	exist, err := kv.Exist(ctx, core.GenerateRBACSecretKey())
	if err != nil {
		log.Errorf(err, "can not get secret info")
		return false, err
	}
	return exist, nil
}
