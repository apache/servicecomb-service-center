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
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-chassis/foundation/tlsutil"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

func ParseSSLCipherSuites(ciphers string, permitTLSCipherSuiteMap map[string]uint16) []uint16 {
	if len(ciphers) == 0 || len(permitTLSCipherSuiteMap) == 0 {
		return nil
	}

	cipherSuiteList := make([]uint16, 0)
	cipherSuiteNameList := strings.Split(ciphers, ",")
	for _, cipherSuiteName := range cipherSuiteNameList {
		cipherSuiteName = strings.TrimSpace(cipherSuiteName)
		if len(cipherSuiteName) == 0 {
			continue
		}

		if cipherSuite, ok := permitTLSCipherSuiteMap[cipherSuiteName]; ok {
			cipherSuiteList = append(cipherSuiteList, cipherSuite)
		} else {
			// 配置算法不存在
			log.Warn(fmt.Sprintf("cipher %s not exist.", cipherSuiteName))
		}
	}

	return cipherSuiteList
}

func ParseDefaultSSLCipherSuites(ciphers string) []uint16 {
	return ParseSSLCipherSuites(ciphers, TLSCipherSuiteMap)
}

func ParseSSLProtocol(sprotocol string) uint16 {
	var result uint16 = tls.VersionTLS12
	if protocol, ok := TLSVersionMap[sprotocol]; ok {
		result = protocol
	} else {
		log.Warn(fmt.Sprintf("invalid ssl minimal version(%s), use default.", sprotocol))
	}

	return result
}

func GetX509CACertPool(caCertFile string) (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		log.Error(fmt.Sprintf("read ca cert file %s failed.", caCertFile), err)
		return nil, err
	}

	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

/**
  verifyPeer    Whether verify client
  supplyCert    Whether send certificate
  verifyCN      Whether verify CommonName
*/
func GetClientTLSConfig(opts ...SSLConfigOption) (tlsConfig *tls.Config, err error) {
	cfg := toSSLConfig(opts...)
	var pool *x509.CertPool = nil
	var certs []tls.Certificate
	if cfg.VerifyPeer {
		pool, err = GetX509CACertPool(cfg.CACertFile)
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.CertFile) > 0 {
		certs, err = tlsutil.LoadTLSCertificate(cfg.CertFile, cfg.KeyFile, cfg.KeyPassphase, nil)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig = &tls.Config{
		RootCAs:            pool,
		Certificates:       certs,
		CipherSuites:       cfg.CipherSuites,
		InsecureSkipVerify: !cfg.VerifyPeer || !cfg.VerifyHostName,
		MinVersion:         cfg.MinVersion,
		MaxVersion:         cfg.MaxVersion,
	}

	return tlsConfig, nil
}

func GetServerTLSConfig(opts ...SSLConfigOption) (tlsConfig *tls.Config, err error) {
	cfg := toSSLConfig(opts...)
	clientAuthMode := tls.NoClientCert
	var pool *x509.CertPool = nil
	if cfg.VerifyPeer {
		pool, err = GetX509CACertPool(cfg.CACertFile)
		if err != nil {
			return nil, err
		}

		clientAuthMode = tls.RequireAndVerifyClientCert
	}

	var certs []tls.Certificate
	if len(cfg.CertFile) > 0 {
		certs, err = tlsutil.LoadTLSCertificate(cfg.CertFile, cfg.KeyFile, cfg.KeyPassphase, nil)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig = &tls.Config{
		ClientCAs:                pool,
		Certificates:             certs,
		CipherSuites:             cfg.CipherSuites,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		PreferServerCipherSuites: true,
		ClientAuth:               clientAuthMode,
		MinVersion:               cfg.MinVersion,
		MaxVersion:               cfg.MaxVersion,
		NextProtos:               []string{"h2", "http/1.1"},
	}

	return tlsConfig, nil
}
