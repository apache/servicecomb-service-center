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

package http

import (
	"crypto/tls"
	"time"
)

type config struct {
	addr              string
	tlsConfig         *tls.Config
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	idleTimeout       time.Duration
	writeTimeout      time.Duration
	maxHeaderBytes    int
	compressed        bool
	compressMinBytes  int
}

// Option to httpserver config
type Option func(*config)

// WithAddr returns address option
func WithAddr(addr string) Option {
	return func(c *config) { c.addr = addr }
}

func WithCompressed(compressed bool) Option {
	return func(c *config) { c.compressed = compressed }
}

func WithCompressMinBytes(compressMinBytes int) Option {
	return func(c *config) { c.compressMinBytes = compressMinBytes }
}

// WithTLSConfig returns tls config option
func WithTLSConfig(conf *tls.Config) Option {
	return func(c *config) { c.tlsConfig = conf }
}

func toHttpServerConfig(ops ...Option) *config {
	conf := &config{}
	for _, op := range ops {
		op(conf)
	}
	return conf
}
