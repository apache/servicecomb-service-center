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
package util

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
)

var connectionPool = make(map[string]*http.Client)

//insecurityConnection define
var insecurityConnection = &http.Client{}

func createConnection(caKey []byte, crt []byte, key []byte) (httpClient *http.Client, err error) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caKey)

	clientCrt, err := tls.X509KeyPair(crt, key)
	if err != nil {
		Logger().Error("X509KeyPair err:", err)
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{clientCrt},
		},
	}
	httpClient = &http.Client{Transport: tr}
	return
}

//NewConnection :create a new connection
func NewConnection(requestPath string, caKey []byte, crt []byte, key []byte) (httpClient *http.Client, err error) {
	requestURL, err := url.Parse(requestPath)
	if err != nil {
		return
	}
	requestHost := requestURL.Host
	if caKey == nil || len(caKey) == 0 { //This is a insecurity connection
		httpClient = insecurityConnection
		connectionPool[requestHost] = insecurityConnection
		return
	}

	//Create a new security connection and update the cache
	httpClient, err = createConnection(caKey, crt, key)
	if err != nil {
		connectionPool[requestHost] = httpClient
	}
	return
}

//GetConnection :get a existing connection
func GetConnection(requestPath string) (httpClient *http.Client, ok bool) {
	requestURL, err := url.Parse(requestPath)
	if err != nil {
		return
	}
	requestHost := requestURL.Host
	existingConn, ok := connectionPool[requestHost]
	if ok { //Get the existing connection in cache
		return existingConn, true
	}
	ok = false
	return
}
