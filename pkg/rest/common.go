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
package rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/ServiceComb/service-center/pkg/tlsutil"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/infra/security"
	"github.com/astaxie/beego"
	"io/ioutil"
	"time"
)

const (
	DEFAULT_TLS_HANDSHAKE_TIMEOUT = 30 * time.Second
	DEFAULT_HTTP_RESPONSE_TIMEOUT = 60 * time.Second

	HTTP_ERROR_STATUS_CODE = 600

	HTTP_METHOD_GET    = "GET"
	HTTP_METHOD_PUT    = "PUT"
	HTTP_METHOD_POST   = "POST"
	HTTP_METHOD_DELETE = "DELETE"

	CTX_RESPONSE      = "_server_response"
	CTX_REQUEST       = "_server_request"
	CTX_MATCH_PATTERN = "_server_match_pattern"
	SERVER_CHAIN_NAME = "_server_chain"
)

func isValidMethod(method string) bool {
	switch method {
	case HTTP_METHOD_GET, HTTP_METHOD_PUT, HTTP_METHOD_POST, HTTP_METHOD_DELETE:
		return true
	default:
		return false
	}
}

func getX509CACertPool() (caCertPool *x509.CertPool, err error) {
	pool := x509.NewCertPool()
	caCertFile := tlsutil.GetServerSSLConfig().CACertFile
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		util.Logger().Errorf(err, "read ca cert file %s failed.", caCertFile)
		return nil, err
	}

	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

func loadTLSCertificate() (tlsCert []tls.Certificate, err error) {
	certFile, keyFile := tlsutil.GetServerSSLConfig().CertFile, tlsutil.GetServerSSLConfig().KeyFile
	passphase := tlsutil.GetServerSSLConfig().KeyPassphase
	plainPassphase, err := security.CipherPlugins[beego.AppConfig.DefaultString("cipher_plugin", "default")]().Decrypt(passphase)
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
		pool, err = getX509CACertPool()
		if err != nil {
			return nil, err
		}
	}

	if supplyCert {
		certs, err = loadTLSCertificate()
		if err != nil {
			return nil, err
		}
	}

	tlsConfig = &tls.Config{
		RootCAs:            pool,
		Certificates:       certs,
		CipherSuites:       tlsutil.GetClientSSLConfig().CipherSuites,
		InsecureSkipVerify: !verifyCN,
		MinVersion:         tlsutil.GetClientSSLConfig().MinVersion,
		MaxVersion:         tlsutil.GetClientSSLConfig().MaxVersion,
	}

	return tlsConfig, nil
}

func GetServerTLSConfig(verifyPeer bool) (tlsConfig *tls.Config, err error) {
	clientAuthMode := tls.NoClientCert
	var pool *x509.CertPool = nil
	if verifyPeer {
		pool, err = getX509CACertPool()
		if err != nil {
			return nil, err
		}

		clientAuthMode = tls.RequireAndVerifyClientCert
	}

	var certs []tls.Certificate
	certs, err = loadTLSCertificate()
	if err != nil {
		return nil, err
	}

	tlsConfig = &tls.Config{
		ClientCAs:                pool,
		Certificates:             certs,
		CipherSuites:             tlsutil.GetServerSSLConfig().CipherSuites,
		PreferServerCipherSuites: true,
		ClientAuth:               clientAuthMode,
		MinVersion:               tlsutil.GetServerSSLConfig().MinVersion,
		MaxVersion:               tlsutil.GetServerSSLConfig().MaxVersion,
	}

	return tlsConfig, nil
}

func GetClient(communiType string) (*HttpClient, error) {
	verifyClient := false
	var err error
	var client *HttpClient
	//client, err = rest.GetHttpsClient(verifyClient)
	if communiType == "https" {
		client, err = GetAnnoHttpsClient(verifyClient)
		if err != nil {
			util.Logger().Error("Create https rest.client failed.", err)
			return nil, err
		}
		return client, nil
	}
	client, err = GetHttpClient(true)
	if err != nil {
		util.Logger().Error("Create http rest.client failed.", err)
		return nil, err
	}
	return client, nil
}
