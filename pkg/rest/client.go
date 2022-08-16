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

package rest

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"

	"github.com/apache/servicecomb-service-center/pkg/buffer"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/foundation/tlsutil"

	"context"
)

var defaultURLClientOption = URLClientOption{
	Compressed:            true,
	VerifyPeer:            true,
	SSLVersion:            tls.VersionTLS12,
	HandshakeTimeout:      10 * time.Second,
	ResponseHeaderTimeout: 30 * time.Second,
	RequestTimeout:        60 * time.Second,
	ConnsPerHost:          DefaultConnPoolPerHostSize,
}

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

type gzipBodyReader struct {
	*gzip.Reader
	Body io.ReadCloser
}

func (w *gzipBodyReader) Close() error {
	w.Reader.Close()
	return w.Body.Close()
}

func NewGZipBodyReader(body io.ReadCloser) (io.ReadCloser, error) {
	reader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	return &gzipBodyReader{reader, body}, nil
}

type URLClient struct {
	*http.Client

	TLS *tls.Config

	Cfg URLClientOption
}

func (client *URLClient) HTTPDoWithContext(ctx context.Context, method string, rawURL string, headers http.Header, body []byte) (resp *http.Response, err error) {
	if strings.HasPrefix(rawURL, "https") {
		if transport, ok := client.Client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = client.TLS
		}
	}

	if headers == nil {
		headers = make(http.Header)
	}

	if _, ok := headers[HeaderHost]; !ok {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		headers.Set(HeaderHost, parsedURL.Host)
	}
	if _, ok := headers[HeaderAccept]; !ok {
		headers.Set(HeaderAccept, AcceptAny)
	}
	if _, ok := headers[HeaderAcceptEncoding]; !ok && client.Cfg.Compressed {
		headers.Set(HeaderAcceptEncoding, "deflate, gzip")
	}

	req, err := http.NewRequest(method, rawURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %s", err.Error())
	}
	req = req.WithContext(ctx)
	req.Header = headers

	DumpRequestOut(req)

	resp, err = client.Client.Do(req)
	if err != nil {
		return nil, err
	}
	DumpResponse(resp)

	switch resp.Header.Get(HeaderContentEncoding) {
	case "gzip":
		reader, err := NewGZipBodyReader(resp.Body)
		if err != nil {
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				log.Error("", err)
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			return nil, err
		}
		resp.Body = reader
	}

	return resp, nil
}

func DumpRequestOut(req *http.Request) {
	if req == nil || !util.StringTRUE(os.Getenv("DEBUG_MODE")) {
		return
	}

	fmt.Println(">", req.URL.String())
	b, _ := httputil.DumpRequestOut(req, true)
	err := buffer.ReadLine(bytes.NewBuffer(b), func(line string) bool {
		fmt.Println(">", line)
		return true
	})
	if err != nil {
		log.Error("", err)
	}
}

func DumpResponse(resp *http.Response) {
	if resp == nil || !util.StringTRUE(os.Getenv("DEBUG_MODE")) {
		return
	}

	b, _ := httputil.DumpResponse(resp, true)
	err := buffer.ReadLine(bytes.NewBuffer(b), func(line string) bool {
		fmt.Println("<", line)
		return true
	})
	if err != nil {
		log.Error("", err)
	}
}

func (client *URLClient) HTTPDo(method string, rawURL string, headers http.Header, body []byte) (resp *http.Response, err error) {
	return client.HTTPDoWithContext(context.Background(), method, rawURL, headers, body)
}

func DefaultURLClientOption() URLClientOption {
	return defaultURLClientOption
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
		opts := append(tlsutil.DefaultClientTLSOptions(),
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
