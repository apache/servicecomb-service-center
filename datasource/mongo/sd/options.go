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

package sd

import (
	"fmt"
	"time"
)

const (
	FirstTimeout         = 2 * time.Second
	DefaultTimeout       = 30 * time.Second
	DefaultCacheInitSize = 100
)

type Options struct {
	// Key is the table to unique specify resource type
	Key      string
	InitSize int
	Timeout  time.Duration
	Period   time.Duration
}

func (options *Options) String() string {
	return fmt.Sprintf("{key: %s, timeout: %s, period: %s}",
		options.Key, options.Timeout, options.Period)
}

func (options *Options) SetTable(key string) *Options {
	options.Key = key
	return options
}

func DefaultOptions() *Options {
	return &Options{
		Key:      "",
		Timeout:  DefaultTimeout,
		Period:   time.Second,
		InitSize: DefaultCacheInitSize,
	}
}
