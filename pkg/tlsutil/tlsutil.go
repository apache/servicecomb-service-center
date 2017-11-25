//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/plugin"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var SERVER_TLS_CIPHER_SUITE_MAP = map[string]uint16{
	"TLS_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
}

var CLIENT_TLS_CIPHER_SUITE_MAP = map[string]uint16{
	"TLS_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":       tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
}

var TLS_VERSION_MAP = map[string]uint16{
	"TLSv1.0": tls.VersionTLS10,
	"TLSv1.1": tls.VersionTLS11,
	"TLSv1.2": tls.VersionTLS12,
}

type SSLConfig struct {
	SSLEnabled   bool
	VerifyClient bool
	CipherSuites []uint16
	MinVersion   uint16
	MaxVersion   uint16
	CACertFile   string
	CertFile     string
	KeyFile      string
	KeyPassphase string
}

var sslServerConfig *SSLConfig = &SSLConfig{
	SSLEnabled:   true,
	VerifyClient: true,
	MinVersion:   tls.VersionTLS12,
	MaxVersion:   tls.VersionTLS12,
	CipherSuites: nil,
	CACertFile:   getSSLPath("trust.cer"),
	CertFile:     getSSLPath("server.cer"),
	KeyFile:      getSSLPath("server_key.pem"),
	KeyPassphase: "",
}

var sslClientConfig *SSLConfig = &SSLConfig{
	SSLEnabled:   true,
	VerifyClient: true,
	CipherSuites: nil,
	MinVersion:   tls.VersionTLS12,
	MaxVersion:   tls.VersionTLS12,
	CACertFile:   getSSLPath("trust.cer"),
	CertFile:     getSSLPath("server.cer"),
	KeyFile:      getSSLPath("server_key.pem"),
	KeyPassphase: "",
}

func getSSLPath(path string) string {
	env := os.Getenv("SSL_ROOT")
	if len(env) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, "etc", "ssl", path)
	}
	return os.ExpandEnv(filepath.Join("$SSL_ROOT", path))
}

func parseSSLCipherSuites(ciphers string, permitTlsCipherSuiteMap map[string]uint16) []uint16 {
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

func parseServerSSLCipherSuites(ciphers string) []uint16 {
	return parseSSLCipherSuites(ciphers, SERVER_TLS_CIPHER_SUITE_MAP)
}

func parseClientSSLCipherSuites(ciphers string) []uint16 {
	return parseSSLCipherSuites(ciphers, CLIENT_TLS_CIPHER_SUITE_MAP)
}

func parseSSLProtocol(sprotocol string) uint16 {
	var result uint16 = tls.VersionTLS12
	if protocol, ok := TLS_VERSION_MAP[sprotocol]; ok {
		result = protocol
	} else {
		util.Logger().Warnf(nil, "invalid ssl minimal version invalid(%s), use default.", sprotocol)
	}

	return result
}

func LoadServerSSLConfig() {
	util.Logger().Debugf("load server ssl configurations.")
	sslServerConfig.SSLEnabled = beego.AppConfig.DefaultInt("ssl_mode", 1) != 0
	sslServerConfig.VerifyClient = beego.AppConfig.DefaultInt("ssl_verify_client", 1) != 0
	sslServerProtocol := beego.AppConfig.DefaultString("ssl_protocols", "TLSv1.2")
	sslServerConfig.MinVersion = parseSSLProtocol(sslServerProtocol)
	sslServerConfig.CipherSuites = parseServerSSLCipherSuites(beego.AppConfig.DefaultString("ssl_ciphers", ""))
	if sslServerConfig.SSLEnabled {
		// 如果配置了SSL模式，SSL参数必须配置
		keyPassphase, err := ioutil.ReadFile(getSSLPath("cert_pwd"))
		if err != nil {
			util.Logger().Warn("read file cert_pwd failed.", err)
		}
		sslServerConfig.KeyPassphase = util.BytesToStringWithNoCopy(keyPassphase)
	}

	util.Logger().Infof("server ssl configs enabled %t, verifyClient %t, minv %#x, ciphers %d, phase %d.",
		sslServerConfig.SSLEnabled,
		sslServerConfig.VerifyClient,
		sslServerConfig.MinVersion,
		len(sslServerConfig.CipherSuites),
		len(sslServerConfig.KeyPassphase))
}

func LoadClientSSLConfig() {
	util.Logger().Debugf("load client ssl configurations.")
	sslClientConfig.SSLEnabled = sslServerConfig.SSLEnabled
	sslClientConfig.VerifyClient = sslServerConfig.VerifyClient
	sslClientProtocol := beego.AppConfig.DefaultString("ssl_client_protocols", "")
	if len(sslClientProtocol) == 0 {
		// 如果未配置，则复用服务端配置
		sslClientConfig.MinVersion = sslServerConfig.MinVersion
	} else {
		sslClientConfig.MinVersion = parseSSLProtocol(sslClientProtocol)
	}

	sslClientCiphers := beego.AppConfig.DefaultString("ssl_client_ciphers", "")
	if len(sslClientCiphers) == 0 {
		// 如果未配置，则复用服务端配置
		sslClientConfig.CipherSuites = sslServerConfig.CipherSuites[:]
	} else {
		sslClientConfig.CipherSuites = parseClientSSLCipherSuites(sslClientCiphers)
	}

	sslClientConfig.KeyPassphase = sslServerConfig.KeyPassphase

	util.Logger().Infof("client ssl configs enabled %t, verifyclient %t, minv %#x, cipers %d, pphase %d.",
		sslClientConfig.SSLEnabled,
		sslClientConfig.VerifyClient,
		sslClientConfig.MinVersion,
		len(sslClientConfig.CipherSuites),
		len(sslClientConfig.KeyPassphase))
}

func GetServerSSLConfig() *SSLConfig {
	return sslServerConfig
}

func GetClientSSLConfig() *SSLConfig {
	return sslClientConfig
}

func GetX509CACertPool() (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCertFile := GetServerSSLConfig().CACertFile
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		util.Logger().Errorf(err, "read ca cert file %s failed.", caCertFile)
		return nil, err
	}

	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

func LoadTLSCertificate() (tlsCert []tls.Certificate, err error) {
	certFile, keyFile := GetServerSSLConfig().CertFile, GetServerSSLConfig().KeyFile
	passphase := GetServerSSLConfig().KeyPassphase
	plainPassphase, err := plugin.Plugins().Cipher().Decrypt(passphase)
	if err != nil {
		util.Logger().Errorf(err, "decrypt ssl passphase(%d) failed.", len(passphase))
		plainPassphase = ""
	}

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
		util.ClearStringMemory(&plainPassphase)
		util.ClearByteMemory(plainPassphaseBytes)
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
func GetClientTLSConfig(verifyPeer bool, supplyCert bool, verifyCN bool) (tlsConfig *tls.Config, err error) {
	var pool *x509.CertPool = nil
	var certs []tls.Certificate
	if verifyPeer {
		pool, err = GetX509CACertPool()
		if err != nil {
			return nil, err
		}
	}

	if supplyCert {
		certs, err = LoadTLSCertificate()
		if err != nil {
			return nil, err
		}
	}

	tlsConfig = &tls.Config{
		RootCAs:            pool,
		Certificates:       certs,
		CipherSuites:       GetClientSSLConfig().CipherSuites,
		InsecureSkipVerify: !verifyCN,
		MinVersion:         GetClientSSLConfig().MinVersion,
		MaxVersion:         GetClientSSLConfig().MaxVersion,
	}

	return tlsConfig, nil
}

func GetServerTLSConfig(verifyPeer bool) (tlsConfig *tls.Config, err error) {
	clientAuthMode := tls.NoClientCert
	var pool *x509.CertPool = nil
	if verifyPeer {
		pool, err = GetX509CACertPool()
		if err != nil {
			return nil, err
		}

		clientAuthMode = tls.RequireAndVerifyClientCert
	}

	var certs []tls.Certificate
	certs, err = LoadTLSCertificate()
	if err != nil {
		return nil, err
	}

	tlsConfig = &tls.Config{
		ClientCAs:                pool,
		Certificates:             certs,
		CipherSuites:             GetServerSSLConfig().CipherSuites,
		PreferServerCipherSuites: true,
		ClientAuth:               clientAuthMode,
		MinVersion:               GetServerSSLConfig().MinVersion,
		MaxVersion:               GetServerSSLConfig().MaxVersion,
	}

	return tlsConfig, nil
}
