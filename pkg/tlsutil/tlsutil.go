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
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
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
			log.Warnf("cipher %s not exist.", cipherSuiteName)
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
		log.Warnf("invalid ssl minimal version(%s), use default.", sprotocol)
	}

	return result
}

func GetX509CACertPool(loader CertLoader) (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCert, err := loader.CADetail()
	if err != nil {
		log.Errorf(err, "read ca cert failed.")
		return nil, err
	}

	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

//CertLoader loads cert info
type CertLoader interface {
	CADetail() ([]byte, error)
	CertDetail() ([]byte, error)
	KeyDetail() ([]byte, error)
	CanLoadCert() bool
}

var defaultLoader = new(fromNone)

type fromNone struct {
}

//CADetail ...
func (fn *fromNone) CADetail() ([]byte, error) {
	return nil, nil
}

//CertDetail ...
func (fn *fromNone) CertDetail() ([]byte, error) {
	return nil, nil
}

//KeyDetail ...
func (fn *fromNone) KeyDetail() ([]byte, error) {
	return nil, nil
}

//CanLoadCert ...
func (fn *fromNone) CanLoadCert() bool {
	return false
}

//FromEnv ...
func FromEnv() *fromEnv {
	return new(fromEnv)
}

type fromEnv struct {
	envCA     string
	envCert   string
	envKey    string
	envKeyPwd string
}

//EnvNameCA ...
func (fe *fromEnv) EnvNameCA(name string) *fromEnv {
	fe.envCA = name
	return fe
}

//EnvNameCert ...
func (fe *fromEnv) EnvNameCert(name string) *fromEnv {
	fe.envCert = name
	return fe
}

//EnvNameKey ...
func (fe *fromEnv) EnvNameKey(name string) *fromEnv {
	fe.envKey = name
	return fe
}

func (fe *fromEnv) getEnv(name string) ([]byte, error) {
	data := os.Getenv(name)
	if len(data) == 0 {
		return nil, fmt.Errorf("%s not exist", name)
	}
	return util.StringToBytesWithNoCopy(data), nil
}

//CADetail ...
func (fe *fromEnv) CADetail() ([]byte, error) {
	return fe.getEnv(fe.envCA)
}

//CertDetail ...
func (fe *fromEnv) CertDetail() ([]byte, error) {
	return fe.getEnv(fe.envCert)
}

//KeyDetail ...
func (fe *fromEnv) KeyDetail() ([]byte, error) {
	return fe.getEnv(fe.envKey)
}

//CanLoadCert ...
func (fe *fromEnv) CanLoadCert() bool {
	return os.Getenv(fe.envCert) != "" &&
		os.Getenv(fe.envKey) != ""
}

//FromFile loads cert info from file
func FromFile() *fromFile {
	return new(fromFile)
}

type fromFile struct {
	caFile   string
	certFile string
	keyFile  string
}

//CaFile ...
func (ff *fromFile) CaFile(name string) *fromFile {
	ff.caFile = name
	return ff
}

//CertFile ...
func (ff *fromFile) CertFile(name string) *fromFile {
	ff.certFile = name
	return ff
}

//KeyFile ...
func (ff *fromFile) KeyFile(name string) *fromFile {
	ff.keyFile = name
	return ff
}

//CADetail ...
func (ff *fromFile) CADetail() ([]byte, error) {
	return ioutil.ReadFile(ff.certFile)
}

//CertDetail ...
func (ff *fromFile) CertDetail() ([]byte, error) {
	return ioutil.ReadFile(ff.certFile)
}

//KeyDetail ...
func (ff *fromFile) KeyDetail() ([]byte, error) {
	return ioutil.ReadFile(ff.keyFile)
}

//CanLoadCert ...
func (ff *fromFile) CanLoadCert() bool {
	return ff.certFile != "" &&
		ff.keyFile != ""
}

func LoadTLSCertificate(loader CertLoader, pwd string) (tlsCert []tls.Certificate, err error) {
	certContent, err := loader.CertDetail()
	if err != nil {
		log.Errorf(err, "read cert file failed.")
		return nil, err
	}

	keyContent, err := loader.KeyDetail()
	if err != nil {
		log.Errorf(err, "read key file failed.")
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyContent)
	if keyBlock == nil {
		log.Errorf(err, "decode key file failed.")
		return nil, err
	}

	if x509.IsEncryptedPEMBlock(keyBlock) {
		plainPwdBytes := util.StringToBytesWithNoCopy(pwd)
		keyData, err := x509.DecryptPEMBlock(keyBlock, plainPwdBytes)
		if err != nil {
			log.Errorf(err, "decrypt key file failed.")
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
		log.Errorf(err, "load X509 key pair failed.")
		return nil, err
	}

	var certs []tls.Certificate
	certs = append(certs, cert)

	return certs, nil
}

func newLoad(cfg *SSLConfig) CertLoader {
	var loader CertLoader = defaultLoader
	if cfg.EnvNameCA != "" {
		return FromEnv().
			EnvNameCA(cfg.EnvNameCA).
			EnvNameCert(cfg.EnvNameCert).
			EnvNameKey(cfg.EnvNameCertKey)
	}
	if len(cfg.CACertFile) == 0 {
		return loader
	}

	return FromFile().
		CaFile(cfg.CACertFile).
		CertFile(cfg.CertFile).
		KeyFile(cfg.KeyFile)
}

/**
  verifyPeer    Whether verify client
  supplyCert    Whether send certificate
  verifyCN      Whether verify CommonName
*/

func GetClientTLSConfig(opts ...SSLConfigOption) (tlsConfig *tls.Config, err error) {
	cfg := toSSLConfig(opts...)
	var (
		pool  *x509.CertPool
		certs []tls.Certificate
	)

	loader := newLoad(&cfg)

	if cfg.VerifyPeer {
		pool, err = GetX509CACertPool(loader)
		if err != nil {
			return nil, err
		}
	}

	if loader.CanLoadCert() {
		certs, err = LoadTLSCertificate(loader, cfg.KeyPassphase)
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
	var pool *x509.CertPool

	loader := newLoad(&cfg)

	if cfg.VerifyPeer {
		pool, err = GetX509CACertPool(loader)
		if err != nil {
			return nil, err
		}

		clientAuthMode = tls.RequireAndVerifyClientCert
	}

	var certs []tls.Certificate
	if loader.CanLoadCert() {
		certs, err = LoadTLSCertificate(loader, cfg.KeyPassphase)
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
