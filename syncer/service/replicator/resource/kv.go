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

package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	"github.com/little-cui/etcdadpt"
)

const (
	KV = "kv"

	ComparableKey = "comparable"
)

const (
	KVKey         = "key"
	KVKeyNonExist = "key not exist in opts"
)

var (
	manager KeyManager

	ErrRecordNonExist = errors.New("record non exist")
)

func NewKV(e *v1sync.Event) Resource {
	r := &kv{
		event:   e,
		manager: keyManage(),
	}
	return r
}

type kv struct {
	event *v1sync.Event
	key   string

	manager         KeyManager
	tombstoneLoader tombstoneLoader

	cur []byte

	defaultFailHandler
}

func (k *kv) LoadCurrentResource(ctx context.Context) *Result {
	key, ok := k.event.Opts[KVKey]
	if !ok {
		return NewResult(Fail, KVKeyNonExist)
	}
	k.key = key

	value, err := k.manager.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrRecordNonExist) {
			return nil
		}
		return FailResult(err)
	}
	k.cur = value
	return nil
}

type Value struct {
	Timestamp int64 `json:"$timestamp"`
}

func (k *kv) getUpdateTime() (int64, error) {
	if k.cur == nil {
		return 0, nil
	}

	comparable, ok := k.event.Opts[ComparableKey]
	if !ok || comparable != "true" {
		return 0, nil
	}

	v := new(Value)
	err := json.Unmarshal(k.cur, v)
	if err != nil {
		log.Warn(fmt.Sprintf("unmarshal kv %s value failed, err %s", k.key, err.Error()))
		return 0, err
	}

	return v.Timestamp, nil
}

func (k *kv) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil:  k.cur != nil,
		event:      k.event,
		updateTime: k.getUpdateTime,
		resourceID: k.key,
	}
	c.tombstoneLoader = c
	if k.tombstoneLoader != nil {
		c.tombstoneLoader = k.tombstoneLoader
	}

	return c.needOperate(ctx)
}

func (k *kv) CreateHandle(ctx context.Context) error {
	return k.manager.Post(ctx, k.key, k.event.Value)
}

func (k *kv) UpdateHandle(ctx context.Context) error {
	return k.manager.Put(ctx, k.key, k.event.Value)
}

func (k *kv) DeleteHandle(ctx context.Context) error {
	return k.manager.Delete(ctx, k.key)
}

var once sync.Once

func keyManage() KeyManager {
	once.Do(InitManager)
	return manager
}

func (k *kv) Operate(ctx context.Context) *Result {
	return newOperator(k).operate(ctx, k.event.Action)
}

type KeyManager interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Post(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}

type etcdManager struct {
}

func InitManager() {
	manager = new(etcdManager)
}

func (e *etcdManager) Get(ctx context.Context, key string) ([]byte, error) {
	r, err := etcdadpt.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, ErrRecordNonExist
	}

	return r.Value, nil
}

func (e etcdManager) Put(ctx context.Context, key string, value []byte) error {
	return etcdadpt.Put(ctx, key, string(value))
}

func (e etcdManager) Post(ctx context.Context, key string, value []byte) error {
	return etcdadpt.Put(ctx, key, string(value))
}

func (e etcdManager) Delete(ctx context.Context, key string) error {
	_, err := etcdadpt.Delete(ctx, key)
	return err
}
