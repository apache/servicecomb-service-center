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

//Package kv supplies kv store
package kv

import (
	"context"
	"errors"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
)

var ErrNotUnique = errors.New("kv result is not unique")

//Put put kv
func Put(ctx context.Context, key string, value string) error {
	_, err := backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue([]byte(value)))
	return err
}

//Put put kv
func PutBytes(ctx context.Context, key string, value []byte) error {
	_, err := backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue(value))
	return err
}

//Get get one kv
func Get(ctx context.Context, key string) (*mvccpb.KeyValue, error) {
	resp, err := backend.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if resp.Count != 1 {
		return nil, ErrNotUnique
	}
	return resp.Kvs[0], err
}

//Get get kv list
func List(ctx context.Context, key string) ([]*mvccpb.KeyValue, int64, error) {
	resp, err := backend.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(key), registry.WithPrefix())
	if err != nil {
		return nil, 0, err
	}
	return resp.Kvs, resp.Count, nil
}

//Exist get one kv, if can not get return false
func Exist(ctx context.Context, key string) (bool, error) {
	resp, err := backend.Registry().Do(ctx, registry.GET,
		registry.WithStrKey(key))
	if err != nil {
		return false, err
	}
	if resp.Count == 0 {
		return false, nil
	}
	return true, nil
}
func Delete(ctx context.Context, key string) (bool, error) {
	resp, err := backend.Registry().Do(ctx, registry.DEL,
		registry.WithStrKey(key))
	if err != nil {
		return false, err
	}
	return resp.Count != 0, nil
}
