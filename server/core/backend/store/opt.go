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

const (
	DEFAULT_MAX_NO_EVENT_INTERVAL     = 1 // TODO it should be set to 1 for prevent etcd data is lost accidentally.
	DEFAULT_LISTWATCH_TIMEOUT         = 30 * time.Second
	DEFAULT_SELF_PRESERVATION_PERCENT = 0.85
)

type KvCacherCfg struct {
	Key                string
	InitSize           int
	NoEventMaxInterval int
	Timeout            time.Duration
	Period             time.Duration
	OnEvent            KvEventFunc
	DeferHander        DeferHandler
}

func (cfg KvCacherCfg) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		cfg.Key, cfg.Timeout, cfg.Period)
}

type KvCacherCfgOption func(*KvCacherCfg)

func WithKey(key string) KvCacherCfgOption {
	return func(cfg *KvCacherCfg) { cfg.Key = key }
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
	return func(cfg *KvCacherCfg) { cfg.DeferHander = h }
}

func DefaultKvCacherConfig() KvCacherCfg {
	return KvCacherCfg{
		Key:                "/",
		Timeout:            DEFAULT_LISTWATCH_TIMEOUT,
		Period:             time.Second,
		NoEventMaxInterval: DEFAULT_MAX_NO_EVENT_INTERVAL,
	}
}
