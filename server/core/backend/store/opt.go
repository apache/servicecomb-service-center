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
package store

import (
	"fmt"
	"time"
)

type KvCacherCfg struct {
	Prefix             string
	InitSize           int
	NoEventMaxInterval int
	Timeout            time.Duration
	Period             time.Duration
	OnEvent            KvEventFunc
	DeferHandler       DeferHandler
}

func (cfg KvCacherCfg) String() string {
	return fmt.Sprintf("{prefix: %s, timeout: %s, period: %s}",
		cfg.Prefix, cfg.Timeout, cfg.Period)
}

type KvCacherCfgOption func(*KvCacherCfg)

func WithPrefix(key string) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.Prefix = key }
}

func WithInitSize(size int) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.InitSize = size }
}

func WithTimeout(ot time.Duration) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.Timeout = ot }
}

func WithPeriod(ot time.Duration) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.Period = ot }
}

func WithEventFunc(f KvEventFunc) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.OnEvent = f }
}

func WithDeferHandler(h DeferHandler) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.DeferHandler = h }
}

func DefaultKvCacherConfig() KvCacherCfg {
	return KvCacherCfg{
		Prefix:             "/",
		Timeout:            DEFAULT_LISTWATCH_TIMEOUT,
		Period:             time.Second,
		NoEventMaxInterval: DEFAULT_MAX_NO_EVENT_INTERVAL,
		InitSize:           DEFAULT_CACHE_INIT_SIZE,
	}
}
