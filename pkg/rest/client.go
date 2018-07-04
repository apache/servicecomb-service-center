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
	"encoding/json"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"time"
)

const (
	DEFAULT_TLS_HANDSHAKE_TIMEOUT = 30 * time.Second
	DEFAULT_HTTP_RESPONSE_TIMEOUT = 10 * time.Second
	DEFAULT_REQUEST_TIMEOUT       = 300 * time.Second

	HTTP_ERROR_STATUS_CODE = 600
)

type HttpClient struct {
	gzip   bool
	Client *http.Client
}

func NewDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
}

func NewTransport() *http.Transport {
	return &http.Transport{
		Dial:                  NewDialer().Dial,
		MaxIdleConnsPerHost:   5,
		ResponseHeaderTimeout: DEFAULT_HTTP_RESPONSE_TIMEOUT,
		TLSHandshakeTimeout:   DEFAULT_TLS_HANDSHAKE_TIMEOUT,
	}
}

/**
  获取普通HTTP客户端
*/

func GetHttpClient(gzip bool) (client *HttpClient, err error) {
	return &HttpClient{
		gzip: gzip,
		Client: &http.Client{
			Transport: NewTransport(),
			Timeout:   DEFAULT_REQUEST_TIMEOUT,
		},
	}, nil
}

func GetClient(urlPath string) (*HttpClient, error) {
	return GetHttpClient(false)
}

func (client *HttpClient) getHeaders(method string, headers map[string]string, body interface{}) map[string]string {
	newHeaders := make(map[string]string)
	if body != nil {
		newHeaders[HEADER_CONTENT_TYPE] = CONTENT_TYPE_JSON
		newHeaders[HEADER_ACCEPT] = ACCEPT_JSON
	}

	if client.gzip {
		newHeaders[HEADER_ACCEPT_ENCODING] = ENCODING_GZIP
		newHeaders[HEADER_CONTENT_ENCODING] = ENCODING_GZIP
	}

	if headers != nil {
		for key, value := range headers {
			newHeaders[key] = value
		}
	}

	return newHeaders
}

func gzipCompress(src []byte) (dst []byte) {
	var byteBuffer bytes.Buffer

	func() {
		gzipWriter := gzip.NewWriter(&byteBuffer)
		defer gzipWriter.Close()
		gzipWriter.Write(src)
	}()

	return byteBuffer.Bytes()
}

func readAndGunzip(reader io.Reader) (dst []byte, err error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		util.Logger().Errorf(err, "duplicate gzip reader failed.")
		return nil, err
	}

	defer gzipReader.Close()
	dst, err = ioutil.ReadAll(gzipReader)
	if err != nil {
		util.Logger().Errorf(err, "read from gzip reader failed.")
		return nil, err
	}

	return dst, nil
}

func (client *HttpClient) httpDo(method string, url string, headers map[string]string, body interface{}) (int, string) {
	status, result := HTTP_ERROR_STATUS_CODE, ""

	var bodyBytes []byte = nil
	var err error = nil
	var bodyReader io.Reader = nil
	if body != nil {
		if headers == nil || len(headers[HEADER_CONTENT_TYPE]) == 0 {
			// 如果请求头未传入Content-Type，则按照json格式进行编码（如果是非json类型，需要自行在headers里指定类型）
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				util.Logger().Errorf(err, "marshal object failed.")
				return status, result
			}
		} else {
			// 如果指定了Content-Type类型，则传入的body必须为byte流
			var ok bool = false
			bodyBytes, ok = body.([]byte)
			if !ok {
				util.Logger().Errorf(nil, "invalid body type '%s'(%s), body must type of byte array if Content-Type specified.", reflect.TypeOf(body), headers[HEADER_CONTENT_TYPE])
				return status, result
			}
		}

		//如果配置了gzip压缩，则对body压缩一次(如果请求头里传入已经gzip压缩了，则不重复压缩)
		if client.gzip && (headers == nil || headers[HEADER_CONTENT_ENCODING] != ENCODING_GZIP) {
			bodyBytes = gzipCompress(bodyBytes)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		util.Logger().Errorf(err, "create request failed.")
		return status, result
	}

	newHeaders := client.getHeaders(method, headers, body)
	for key, value := range newHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		util.Logger().Errorf(err, "invoke request failed.")
		return status, result
	}

	defer resp.Body.Close()
	status = resp.StatusCode
	var respBody []byte
	if resp.Header.Get(HEADER_CONTENT_ENCODING) == ENCODING_GZIP {
		// 如果响应头里包含了响应消息的压缩格式为gzip，则在返回前先解压缩
		respBody, _ = readAndGunzip(resp.Body)
	} else {
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	result = util.BytesToStringWithNoCopy(respBody)

	return status, result
}

func (client *HttpClient) HttpDo(method string, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte = nil
	var err error = nil
	var bodyReader io.Reader = nil
	if body != nil {
		if headers == nil || len(headers[HEADER_CONTENT_TYPE]) == 0 {
			// 如果请求头未传入Conent-Type，则按照json格式进行编码（如果是非json类型，需要自行在headers里指定类型）
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				util.Logger().Errorf(err, "marshal object failed.")
				return nil, err
			}
		} else {
			// 如果指定了Content-Type类型，则传入的body必须为byte流
			var ok bool = false
			bodyBytes, ok = body.([]byte)
			if !ok {
				err := fmt.Errorf("invalid body type '%s'(%s), body must type of byte array if Content-Type specified.",
					reflect.TypeOf(body), headers[HEADER_CONTENT_TYPE])
				util.Logger().Errorf(err, "")
				return nil, err
			}
		}

		//如果配置了gzip压缩，则对body压缩一次(如果请求头里传入已经gzip压缩了，则不重复压缩)
		if client.gzip && (headers == nil || headers[HEADER_CONTENT_ENCODING] != ENCODING_GZIP) {
			bodyBytes = gzipCompress(bodyBytes)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		util.Logger().Errorf(err, "create request failed.")
		return nil, err
	}

	newHeaders := client.getHeaders(method, headers, body)
	for key, value := range newHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		util.Logger().Errorf(err, "Request -----> %s failed.", url)
		return resp, err
	}
	return resp, err
}

func (client *HttpClient) Get(url string, headers map[string]string) (int, string) {
	return client.httpDo(http.MethodGet, url, headers, nil)
}

func (client *HttpClient) Put(url string, headers map[string]string, body interface{}) (int, string) {
	return client.httpDo(http.MethodPut, url, headers, body)
}

func (client *HttpClient) Post(url string, headers map[string]string, body interface{}) (int, string) {
	return client.httpDo(http.MethodPost, url, headers, body)
}

func (client *HttpClient) Delete(url string, headers map[string]string) (int, string) {
	return client.httpDo(http.MethodDelete, url, headers, nil)
}

func (client *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.Client.Do(req)
}
