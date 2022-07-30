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

package datasource

import (
	"crypto/tls"
	"time"

	"github.com/go-chassis/openlog"
)

//Options contains configuration for plugins
type Options struct {
	// Kind plugin kind, can be 'etcd' or 'mongo'
	Kind string `json:"-"`
	// Logger logger for adapter, by default use openlog.GetLogger()
	Logger     openlog.Logger `json:"-"`
	SslEnabled bool           `json:"-"`
	TLSConfig  *tls.Config    `json:"-"`
	// ErrorFunc called when connection error occurs
	ErrorFunc func(err error) `json:"-"`
	// ConnectedFunc called when connected
	ConnectedFunc func() `json:"-"`

	// CompactInterval optional, compact backend
	CompactInterval   time.Duration `json:"-"`
	CompactIndexDelta int64         `json:"-"`

	EnableCache bool
	// InstanceTTL: the default ttl of instance lease
	InstanceTTL        int64
	InstanceProperties map[string]string
}
