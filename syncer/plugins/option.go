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

package plugins

import (
	"github.com/apache/servicecomb-service-center/client"
)

type scOption struct {
	endpoints     []string
	tlsEnabled    bool
	tlsVerifyPeer bool
	tlsPassphrase string
	tlsCAFile     string
	tlsCertFile   string
	tlsKeyFile    string
}

type SCConfigOption func(*scOption)

func WithEndpoints(endpoints []string) SCConfigOption {
	return func(c *scOption) { c.endpoints = endpoints }
}

func WithTLSEnabled(tlsEnabled bool) SCConfigOption {
	return func(c *scOption) { c.tlsEnabled = tlsEnabled }
}

func WithTLSVerifyPeer(tlsVerifyPeer bool) SCConfigOption {
	return func(c *scOption) { c.tlsVerifyPeer = tlsVerifyPeer }
}

func WithTLSPassphrase(tlsPassphrase string) SCConfigOption {
	return func(c *scOption) { c.tlsPassphrase = tlsPassphrase }
}

func WithTLSCAFile(tlsCAFile string) SCConfigOption {
	return func(c *scOption) { c.tlsCAFile = tlsCAFile }
}

func WithTLSCertFile(tlsCertFile string) SCConfigOption {
	return func(c *scOption) { c.tlsCertFile = tlsCertFile }
}

func WithTLSKeyFile(tlsKeyFile string) SCConfigOption {
	return func(c *scOption) { c.tlsKeyFile = tlsKeyFile }
}

func ToSCConfig(opts ...SCConfigOption) client.Config {
	op := scOption{}
	for _, opt := range opts {
		opt(&op)
	}
	conf := client.Config{Endpoints: op.endpoints}
	if op.tlsEnabled {
		conf.VerifyPeer = op.tlsVerifyPeer
		conf.CAFile = op.tlsCAFile
		conf.CertFile = op.tlsCertFile
		conf.CertKeyFile = op.tlsKeyFile
		conf.CertKeyPWD = op.tlsPassphrase
	}
	return conf
}
