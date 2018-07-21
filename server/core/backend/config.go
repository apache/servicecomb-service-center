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
package backend

import (
	"fmt"
	"golang.org/x/net/context"
	"time"
)

type Config struct {
	Prefix         string
	InitSize       int
	NoEventPeriods int
	Timeout        time.Duration
	Period         time.Duration
	DeferHandler   DeferHandler
	OnEvent        KvEventFunc
	Parser         *Parser
}

func (cfg *Config) String() string {
	return fmt.Sprintf("{prefix: %s, timeout: %s, period: %s}",
		cfg.Prefix, cfg.Timeout, cfg.Period)
}

func (cfg *Config) WithPrefix(key string) *Config {
	cfg.Prefix = key
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

func (cfg *Config) WithNoEventPeriods(p int) *Config {
	cfg.NoEventPeriods = p
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

func (cfg *Config) WithParser(parser *Parser) *Config {
	cfg.Parser = parser
	return cfg
}

func Configure() *Config {
	return &Config{
		Prefix:         "/",
		Timeout:        DEFAULT_LISTWATCH_TIMEOUT,
		Period:         time.Second,
		NoEventPeriods: DEFAULT_MAX_NO_EVENT_INTERVAL,
		InitSize:       DEFAULT_CACHE_INIT_SIZE,
		Parser:         BytesParser,
	}
}

type ListWatchConfig struct {
	Timeout time.Duration
	Context context.Context
}

func (lo *ListWatchConfig) String() string {
	return fmt.Sprintf("{timeout: %s}", lo.Timeout)
}
