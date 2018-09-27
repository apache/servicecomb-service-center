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

package rest

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var defaultURLClientOption = URLClientOption{
	Compressed:            true,
	VerifyPeer:            true,
	SSLVersion:            tls.VersionTLS12,
	HandshakeTimeout:      30 * time.Second,
	ResponseHeaderTimeout: 180 * time.Second,
	RequestTimeout:        300 * time.Second,
	ConnsPerHost:          DEFAULT_CONN_POOL_PER_HOST_SIZE,
}

var defaultClientTLSOptions = tlsutil.DefaultClientTLSOptions()

type URLClientOption struct {
	SSLEnabled            bool
	Compressed            bool
	VerifyPeer            bool
	CAFile                string
	CertFile              string
	CertKeyFile           string
	CertKeyPWD            string
	SSLVersion            uint16
	HandshakeTimeout      time.Duration
	ResponseHeaderTimeout time.Duration
	RequestTimeout        time.Duration
	ConnsPerHost          int
}

type URLClient struct {
	*http.Client

	TLS *tls.Config

	Request *http.Request

	Cfg URLClientOption
}

func (client *URLClient) HttpDo(method string, rawURL string, headers http.Header, body []byte) (resp *http.Response, err error) {
	if strings.HasPrefix(rawURL, "https") {
		if transport, ok := client.Client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = client.TLS
		}
	}

	if headers == nil {
		headers = make(http.Header)
	}

	if _, ok := headers["Host"]; !ok {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		headers.Set("Host", parsedURL.Host)
	}
	if _, ok := headers["Accept"]; !ok {
		headers.Set("Accept", "*/*")
	}
	if _, ok := headers["Accept-Encoding"]; !ok && client.Cfg.Compressed {
		headers.Set("Accept-Encoding", "deflate, gzip")
	}

	req, err := http.NewRequest(method, rawURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("create request failed: %s", err.Error()))
	}
	client.Request = req

	req.Header = headers

	resp, err = client.Client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("invoke request failed: %s", err.Error()))
	}

	if os.Getenv("DEBUG_MODE") == "1" {
		fmt.Println("--- BEGIN ---")
		fmt.Printf("> %s %s %s\n", client.Request.Method, client.Request.URL.RequestURI(), client.Request.Proto)
		for key, header := range client.Request.Header {
			for _, value := range header {
				fmt.Printf("> %s: %s\n", key, value)
			}
		}
		fmt.Println(">")
		fmt.Println(util.BytesToStringWithNoCopy(body))
		fmt.Printf("< %s %s\n", resp.Proto, resp.Status)
		for key, header := range resp.Header {
			for _, value := range header {
				fmt.Printf("< %s: %s\n", key, value)
			}
		}
		fmt.Println("<")
		fmt.Println("--- END ---")
	}
	return resp, nil
}

func setOptionDefaultValue(o *URLClientOption) URLClientOption {
	if o == nil {
		return defaultURLClientOption
	}

	option := *o
	if option.RequestTimeout <= 0 {
		option.RequestTimeout = defaultURLClientOption.RequestTimeout
	}
	if option.HandshakeTimeout <= 0 {
		option.HandshakeTimeout = defaultURLClientOption.HandshakeTimeout
	}
	if option.ResponseHeaderTimeout <= 0 {
		option.ResponseHeaderTimeout = defaultURLClientOption.ResponseHeaderTimeout
	}
	if option.SSLVersion == 0 {
		option.SSLVersion = defaultURLClientOption.SSLVersion
	}
	return option
}

func GetURLClient(o URLClientOption) (client *URLClient, err error) {
	option := setOptionDefaultValue(&o)
	client = &URLClient{
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   option.ConnsPerHost,
				TLSHandshakeTimeout:   option.HandshakeTimeout,
				ResponseHeaderTimeout: option.ResponseHeaderTimeout,
				DisableCompression:    !option.Compressed,
			},
			Timeout: option.RequestTimeout,
		},
		Cfg: option,
	}

	if option.SSLEnabled {
		opts := append(defaultClientTLSOptions,
			tlsutil.WithVerifyPeer(option.VerifyPeer),
			tlsutil.WithCA(option.CAFile),
			tlsutil.WithCert(option.CertFile),
			tlsutil.WithKey(option.CertKeyFile),
			tlsutil.WithKeyPass(option.CertKeyPWD))

		client.TLS, err = tlsutil.GetClientTLSConfig(opts...)
		if err != nil {
			return nil, err
		}
	}
	return
}
