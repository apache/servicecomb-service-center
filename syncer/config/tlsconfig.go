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

package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
)

const (
	dirName    = "certs"
	caCert     = "trust.cer"
	serverCert = "server.cer"
	serverKey  = "server_key.pem"
)

// TLSConfig tls configuration
type TLSConfig struct {
	Enabled         bool        `yaml:"enabled"`
	VerifyPeer      bool        `yaml:"verify_peer"`
	MinVersion      string      `yaml:"min_version"`
	Passphrase      string      `yaml:"passphrase"`
	CAFile          string      `yaml:"ca_file"`
	CertFile        string      `yaml:"cert_file"`
	KeyFile         string      `yaml:"key_file"`
	Ciphers         []string    `yaml:"ciphers"`
	clientTlsConfig *tls.Config `yaml:"-"`
	serverTlsConfig *tls.Config `yaml:"-"`
	mux             sync.Mutex  `yaml:"-"`
}

// DefaultTLSConfig returns default tls configuration
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		Enabled: false,
	}
}

// NewTLSConfig returns tls configuration by server
func NewTLSConfig(server string) *TLSConfig {
	return &TLSConfig{
		Enabled:    true,
		VerifyPeer: true,
		CAFile:     pathFromSSLEnvOrDefault(server, caCert),
		CertFile:   pathFromSSLEnvOrDefault(server, serverCert),
		KeyFile:    pathFromSSLEnvOrDefault(server, serverKey),
	}
}

// Merge other tls configuration into the current configuration
func (t *TLSConfig) Merge(server string, other *TLSConfig) {
	if !other.Enabled {
		return
	}
	t.Enabled = other.Enabled
	t.VerifyPeer = other.VerifyPeer
	t.MinVersion = other.MinVersion

	if other.CAFile == "" && t.CAFile == "" {
		other.CAFile = pathFromSSLEnvOrDefault(server, caCert)
	}
	t.CAFile = other.CAFile

	if other.CertFile == "" && t.CertFile == "" {
		other.CertFile = pathFromSSLEnvOrDefault(server, serverCert)
	}
	t.CertFile = other.CertFile

	if other.KeyFile == "" && t.KeyFile == "" {
		other.KeyFile = pathFromSSLEnvOrDefault(server, serverKey)
	}
	t.KeyFile = other.KeyFile
}

// Verify the tls configuration
func (t *TLSConfig) Verify() (err error) {
	if !t.Enabled {
		return
	}

	if t.CAFile == "" || !utils.IsFileExist(t.CAFile) {
		err = fmt.Errorf("tls ca file '%s' is not found", t.CAFile)
	}

	if err == nil && t.CertFile == "" || !utils.IsFileExist(t.CertFile) {
		err = fmt.Errorf("tls cert file '%s' is not found", t.CertFile)
	}

	if err == nil && t.KeyFile == "" || !utils.IsFileExist(t.KeyFile) {
		err = fmt.Errorf("tls key file '%s' is not found", t.KeyFile)
	}

	if err == nil {
		for _, cipher := range t.Ciphers {
			if _, ok := tlsutil.TLS_CIPHER_SUITE_MAP[cipher]; !ok {
				err = fmt.Errorf("cipher %s not exist", cipher)
				break
			}
		}
	}

	if err != nil {
		log.Error("verify tls configuration failed", err)
	}
	return
}

// ClientTlsConfig get the tls.config of client
func (t *TLSConfig) ClientTlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	t.mux.Lock()
	defer t.mux.Unlock()
	if t.clientTlsConfig != nil {
		return t.clientTlsConfig, nil
	}

	opts := append(tlsutil.DefaultClientTLSOptions(), t.toOptions()...)
	conf, err := tlsutil.GetClientTLSConfig(opts...)
	if err != nil {
		log.Error("get client tls config failed", err)
	}
	t.clientTlsConfig = conf
	return conf, err
}

// ServerTlsConfig get the tls.config of server
func (t *TLSConfig) ServerTlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	if t.serverTlsConfig != nil {
		return t.serverTlsConfig, nil
	}

	opts := append(tlsutil.DefaultServerTLSOptions(), t.toOptions()...)
	conf, err := tlsutil.GetServerTLSConfig(opts...)
	if err != nil {
		log.Error("get server tls config failed", err)
	}
	t.serverTlsConfig = conf
	return conf, err
}

func (t *TLSConfig) toOptions() []tlsutil.SSLConfigOption {
	return []tlsutil.SSLConfigOption{
		tlsutil.WithVerifyPeer(t.VerifyPeer),
		tlsutil.WithVersion(tlsutil.ParseSSLProtocol(t.MinVersion), tls.VersionTLS12),
		tlsutil.WithCipherSuits(
			tlsutil.ParseDefaultSSLCipherSuites(strings.Join(t.Ciphers, ","))),
		tlsutil.WithKeyPass(t.Passphrase),
		tlsutil.WithCA(t.CAFile),
		tlsutil.WithCert(t.CertFile),
		tlsutil.WithKey(t.KeyFile),
	}
}

func pathFromSSLEnvOrDefault(server, path string) string {
	env := os.Getenv("SSL_ROOT")
	if len(env) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, dirName, server, path)
	}
	return os.ExpandEnv(filepath.Join("$SSL_ROOT", server, path))
}
