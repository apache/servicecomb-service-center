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

package serf

import (
	"io"

	"github.com/hashicorp/serf/serf"
)

// Option func
type Option func(*serf.Config)

// WithNode returns node option
func WithNode(nodeName string) Option {
	return func(c *serf.Config) { c.NodeName = nodeName }
}

// WithTags returns tags option
func WithTags(tags map[string]string) Option {
	return func(c *serf.Config) { c.Tags = tags }
}

// WithAddTag returns add tag option
func WithAddTag(key, val string) Option {
	return func(c *serf.Config) { c.Tags[key] = val }
}

// WithBindAddr returns bind addr option
func WithBindAddr(addr string) Option {
	return func(c *serf.Config) { c.MemberlistConfig.BindAddr = addr }
}

// WithBindPort returns bind port option
func WithBindPort(port int) Option {
	return func(c *serf.Config) { c.MemberlistConfig.BindPort = port }
}

// WithAdvertiseAddr returns advertise addr option
func WithAdvertiseAddr(addr string) Option {
	return func(c *serf.Config) { c.MemberlistConfig.AdvertiseAddr = addr }
}

// WithAdvertisePort returns advertise port option
func WithAdvertisePort(port int) Option {
	return func(c *serf.Config) { c.MemberlistConfig.AdvertisePort = port }
}

// WithEnableCompression returns enable compression option
func WithEnableCompression(enable bool) Option {
	return func(c *serf.Config) { c.MemberlistConfig.EnableCompression = enable }
}

// WithSecretKey returns secret key option
func WithSecretKey(secretKey []byte) Option {
	return func(c *serf.Config) { c.MemberlistConfig.SecretKey = secretKey }
}

// WithLogOutput returns log output option
func WithLogOutput(logOutput io.Writer) Option {
	return func(c *serf.Config) {
		c.LogOutput = logOutput
		c.MemberlistConfig.LogOutput = logOutput
	}
}
