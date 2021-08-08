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

package state

import (
	"time"

	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/parser"
)

type Option func(options *kvstore.Options)

func WithPrefix(key string) Option {
	return func(opts *kvstore.Options) {
		opts.Key = key
	}
}

func WithInitSize(size int) Option {
	return func(opts *kvstore.Options) {
		opts.InitSize = size
	}
}

func WithTimeout(ot time.Duration) Option {
	return func(opts *kvstore.Options) {
		opts.Timeout = ot
	}
}

func WithPeriod(ot time.Duration) Option {
	return func(opts *kvstore.Options) {
		opts.Period = ot
	}
}

func WithDeferHandler(h kvstore.DeferHandler) Option {
	return func(opts *kvstore.Options) {
		opts.DeferHandler = h
	}
}

func WithParser(parser parser.Parser) Option {
	return func(opts *kvstore.Options) {
		opts.Parser = parser
	}
}

func ToOptions(opts ...Option) kvstore.Options {
	def := kvstore.NewOptions()
	config := *def
	for _, opt := range opts {
		opt(&config)
	}
	return config
}
