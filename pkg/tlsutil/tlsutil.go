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
package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"io/ioutil"
	"strings"
)

func ParseSSLCipherSuites(ciphers string, permitTlsCipherSuiteMap map[string]uint16) []uint16 {
	if len(ciphers) == 0 || len(permitTlsCipherSuiteMap) == 0 {
		return nil
	}

	cipherSuiteList := make([]uint16, 0)
	cipherSuiteNameList := strings.Split(ciphers, ",")
	for _, cipherSuiteName := range cipherSuiteNameList {
		cipherSuiteName = strings.TrimSpace(cipherSuiteName)
		if len(cipherSuiteName) == 0 {
			continue
		}

		if cipherSuite, ok := permitTlsCipherSuiteMap[cipherSuiteName]; ok {
			cipherSuiteList = append(cipherSuiteList, cipherSuite)
		} else {
			// 配置算法不存在
			util.Logger().Warnf(nil, "cipher %s not exist.", cipherSuiteName)
		}
	}

	return cipherSuiteList
}

func ParseDefaultSSLCipherSuites(ciphers string) []uint16 {
	return ParseSSLCipherSuites(ciphers, TLS_CIPHER_SUITE_MAP)
}

func ParseSSLProtocol(sprotocol string) uint16 {
	var result uint16 = tls.VersionTLS12
	if protocol, ok := TLS_VERSION_MAP[sprotocol]; ok {
		result = protocol
	} else {
		util.Logger().Warnf(nil, "invalid ssl minimal version(%s), use default.", sprotocol)
	}

	return result
}

func GetX509CACertPool(caCertFile string) (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		util.Logger().Errorf(err, "read ca cert file %s failed.", caCertFile)
		return nil, err
	}

	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

func LoadTLSCertificate(certFile, keyFile, plainPassphase string) (tlsCert []tls.Certificate, err error) {
	certContent, err := ioutil.ReadFile(certFile)
	if err != nil {
		util.Logger().Errorf(err, "read cert file %s failed.", certFile)
		return nil, err
	}

	keyContent, err := ioutil.ReadFile(keyFile)
	if err != nil {
		util.Logger().Errorf(err, "read key file %s failed.", keyFile)
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyContent)
	if keyBlock == nil {
		util.Logger().Errorf(err, "decode key file %s failed.", keyFile)
		return nil, err
	}

	if x509.IsEncryptedPEMBlock(keyBlock) {
		plainPassphaseBytes := util.StringToBytesWithNoCopy(plainPassphase)
		keyData, err := x509.DecryptPEMBlock(keyBlock, plainPassphaseBytes)
		if err != nil {
			util.Logger().Errorf(err, "decrypt key file %s failed.", keyFile)
			return nil, err
		}

		// 解密成功，重新编码为PEM格式的文件
		plainKeyBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: keyData,
		}

		keyContent = pem.EncodeToMemory(plainKeyBlock)
	}

	cert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		util.Logger().Errorf(err, "load X509 key pair from cert file %s with key file %s failed.", certFile, keyFile)
		return nil, err
	}

	var certs []tls.Certificate
	certs = append(certs, cert)

	return certs, nil
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
		certs, err = LoadTLSCertificate(cfg.CertFile, cfg.KeyFile, cfg.KeyPassphase)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig = &tls.Config{
		RootCAs:            pool,
		Certificates:       certs,
		CipherSuites:       cfg.CipherSuites,
		InsecureSkipVerify: !cfg.VerifyHostName,
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
		certs, err = LoadTLSCertificate(cfg.CertFile, cfg.KeyFile, cfg.KeyPassphase)
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
