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

package kvstore

import (
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
)

type Options struct {
	// Key is the prefix to unique specify resource type
	Key          string
	InitSize     int
	Timeout      time.Duration
	Period       time.Duration
	DeferHandler DeferHandler
	OnEvent      EventFunc
	Parser       parser.Parser
}

func (opts *Options) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		opts.Key, opts.Timeout, opts.Period)
}

func (opts *Options) WithPrefix(key string) *Options {
	opts.Key = key
	return opts
}

func (opts *Options) WithInitSize(size int) *Options {
	opts.InitSize = size
	return opts
}

func (opts *Options) WithTimeout(ot time.Duration) *Options {
	opts.Timeout = ot
	return opts
}

func (opts *Options) WithPeriod(ot time.Duration) *Options {
	opts.Period = ot
	return opts
}

func (opts *Options) WithDeferHandler(h DeferHandler) *Options {
	opts.DeferHandler = h
	return opts
}

func (opts *Options) WithEventFunc(f EventFunc) *Options {
	opts.OnEvent = f
	return opts
}

func (opts *Options) AppendEventFunc(f EventFunc) *Options {
	if prev := opts.OnEvent; prev != nil {
		next := f
		f = func(evt Event) {
			prev(evt)
			next(evt)
		}
	}
	opts.OnEvent = f
	return opts
}

func (opts *Options) WithParser(parser parser.Parser) *Options {
	opts.Parser = parser
	return opts
}

func NewOptions() *Options {
	return &Options{
		Key:      "/",
		Timeout:  DefaultTimeout,
		Period:   time.Second,
		InitSize: DefaultCacheInitSize,
		Parser:   parser.BytesParser,
	}
}
