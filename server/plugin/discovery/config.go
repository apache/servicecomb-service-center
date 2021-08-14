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
package discovery

import (
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/server/core/proto"
)

type Config struct {
	// Key is the prefix to unique specify resource type
	Key          string
	InitSize     int
	Timeout      time.Duration
	Period       time.Duration
	DeferHandler DeferHandler
	OnEvent      KvEventFunc
	Parser       proto.Parser
	Lease        bool
}

func (cfg *Config) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s, lease: %v}",
		cfg.Key, cfg.Timeout, cfg.Period, cfg.Lease)
}

func (cfg *Config) WithPrefix(key string) *Config {
	cfg.Key = key
	return cfg
}

func (cfg *Config) WithInitSize(size int) *Config {
	cfg.InitSize = size
	return cfg
}

func (cfg *Config) WithTimeout(ot time.Duration) *Config {
	cfg.Timeout = ot
	return cfg
}

func (cfg *Config) WithPeriod(ot time.Duration) *Config {
	cfg.Period = ot
	return cfg
}

func (cfg *Config) WithDeferHandler(h DeferHandler) *Config {
	cfg.DeferHandler = h
	return cfg
}

func (cfg *Config) WithEventFunc(f KvEventFunc) *Config {
	cfg.OnEvent = f
	return cfg
}

func (cfg *Config) AppendEventFunc(f KvEventFunc) *Config {
	if prev := cfg.OnEvent; prev != nil {
		next := f
		f = func(evt KvEvent) {
			prev(evt)
			next(evt)
		}
	}
	cfg.OnEvent = f
	return cfg
}

func (cfg *Config) WithParser(parser proto.Parser) *Config {
	cfg.Parser = parser
	return cfg
}

func (cfg *Config) WithLease() *Config {
	cfg.Lease = true
	return cfg
}

func Configure() *Config {
	return &Config{
		Key:      "/",
		Timeout:  DefaultTimeout,
		Period:   time.Second,
		InitSize: DefaultCacheInitSize,
		Parser:   proto.BytesParser,
	}
}
