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
package common

import (
	"crypto/tls"
	"github.com/ServiceComb/service-center/util"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var TLS_CIPHER_SUITE_MAP = map[string]uint16{
	"TLS_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
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

func parseSSLCipherSuites(ciphers string) []uint16 {
	cipherSuiteList := make([]uint16, 0)
	cipherSuiteNameList := strings.Split(ciphers, ",")
	for _, cipherSuiteName := range cipherSuiteNameList {
		cipherSuiteName = strings.TrimSpace(cipherSuiteName)
		if len(cipherSuiteName) == 0 {
			continue
		}

		if cipherSuite, ok := TLS_CIPHER_SUITE_MAP[cipherSuiteName]; ok {
			cipherSuiteList = append(cipherSuiteList, cipherSuite)
		} else {
			// 配置算法不存在
			util.LOGGER.Warnf(nil, "cipher %s not exist.", cipherSuiteName)
		}
	}

	return cipherSuiteList
}

func parseSSLProtocol(sprotocol string) uint16 {
	var result uint16 = tls.VersionTLS12
	if protocol, ok := TLS_VERSION_MAP[sprotocol]; ok {
		result = protocol
	} else {
		util.LOGGER.Warnf(nil, "invalid ssl minimal version invalid(%s), use default.", sprotocol)
	}

	return result
}

func loadServerSSLConfig() {
	util.LOGGER.Warnf(nil, "load server ssl configurations.")
	sslServerConfig.SSLEnabled = beego.AppConfig.DefaultInt("ssl_mode", 1) != 0
	sslServerConfig.VerifyClient = beego.AppConfig.DefaultInt("ssl_verify_client", 1) != 0
	sslServerProtocol := beego.AppConfig.DefaultString("ssl_protocols", "TLSv1.2")
	sslServerConfig.MinVersion = parseSSLProtocol(sslServerProtocol)
	sslServerConfig.CipherSuites = parseSSLCipherSuites(beego.AppConfig.DefaultString("ssl_ciphers", ""))
	if sslServerConfig.SSLEnabled {
		// 如果配置了SSL模式，SSL参数必须配置
		keyPassphase, err := ioutil.ReadFile(getSSLPath("cert_pwd"))
		if err != nil {
			util.LOGGER.Error("read file cert_pwd failed.", err)
			return
		}
		sslServerConfig.KeyPassphase = string(keyPassphase)
	}

	util.LOGGER.Warnf(nil, "server ssl configs enabled %t, verifyclient %t, minv %#x, cipers %d, pphase %d.",
		sslServerConfig.SSLEnabled,
		sslServerConfig.VerifyClient,
		sslServerConfig.MinVersion,
		len(sslServerConfig.CipherSuites),
		len(sslServerConfig.KeyPassphase))
}

func loadClientSSLConfig() {
	util.LOGGER.Warnf(nil, "load client ssl configurations.")
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
		sslClientConfig.CipherSuites = parseSSLCipherSuites(sslClientCiphers)
	}

	sslClientConfig.KeyPassphase = sslServerConfig.KeyPassphase

	util.LOGGER.Warnf(nil, "client ssl configs enabled %t, verifyclient %t, minv %#x, cipers %d, pphase %d.",
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
