// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tlsutil

import (
	"crypto/tls"
)

type SSLConfig struct {
	VerifyPeer     bool
	VerifyHostName bool
	CipherSuites   []uint16
	MinVersion     uint16
	MaxVersion     uint16
	CACertFile     string
	CertFile       string
	KeyFile        string
	KeyPassphase   string

	EnvNameCA      string
	EnvNameCert    string
	EnvNameCertKey string
}

type SSLConfigOption func(*SSLConfig)

func WithVerifyPeer(b bool) SSLConfigOption      { return func(c *SSLConfig) { c.VerifyPeer = b } }
func WithVerifyHostName(b bool) SSLConfigOption  { return func(c *SSLConfig) { c.VerifyHostName = b } }
func WithCipherSuits(s []uint16) SSLConfigOption { return func(c *SSLConfig) { c.CipherSuites = s } }
func WithVersion(min, max uint16) SSLConfigOption {
	return func(c *SSLConfig) { c.MinVersion, c.MaxVersion = min, max }
}

//WithEnvNameCA sets env name of ca
func WithEnvNameCA(n string) SSLConfigOption { return func(c *SSLConfig) { c.EnvNameCA = n } }

//WithEnvNameCert sets env name of cert
func WithEnvNameCert(n string) SSLConfigOption { return func(c *SSLConfig) { c.EnvNameCert = n } }

//WithEnvNameCertKey sets env name of cert pwd
func WithEnvNameCertKey(n string) SSLConfigOption { return func(c *SSLConfig) { c.EnvNameCertKey = n } }

func WithCA(f string) SSLConfigOption      { return func(c *SSLConfig) { c.CACertFile = f } }
func WithCert(f string) SSLConfigOption    { return func(c *SSLConfig) { c.CertFile = f } }
func WithKey(k string) SSLConfigOption     { return func(c *SSLConfig) { c.KeyFile = k } }
func WithKeyPass(p string) SSLConfigOption { return func(c *SSLConfig) { c.KeyPassphase = p } }

func toSSLConfig(opts ...SSLConfigOption) (op SSLConfig) {
	for _, opt := range opts {
		opt(&op)
	}
	return
}

func DefaultClientTLSOptions() []SSLConfigOption {
	return []SSLConfigOption{
		WithVerifyPeer(true),
		WithVerifyHostName(true),
		WithVersion(tls.VersionTLS12, MaxSupportedTLSVersion),
	}
}

func DefaultServerTLSOptions() []SSLConfigOption {
	return []SSLConfigOption{
		WithVerifyPeer(true),
		WithVersion(tls.VersionTLS12, MaxSupportedTLSVersion),
		WithCipherSuits(TLSCipherSuits()),
	}
}
