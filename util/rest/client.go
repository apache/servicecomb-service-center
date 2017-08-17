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
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/errors"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"time"
)

type HttpClient struct {
	gzip   bool
	client *http.Client
}

func getTLSTransport(verifyPeer bool, supplyCert bool, verifyCN bool) (transport *http.Transport, err error) {
	tlsConfig, err := GetClientTLSConfig(verifyPeer, supplyCert, verifyCN)
	if err != nil {
		return nil, err
	}

	transport = &http.Transport{
		TLSClientConfig:       tlsConfig,
		TLSHandshakeTimeout:   DEFAULT_TLS_HANDSHAKE_TIMEOUT,
		ResponseHeaderTimeout: DEFAULT_HTTP_RESPONSE_TIMEOUT,
	}

	return transport, nil
}

/**
  获取普通HTTP客户端
*/
var httpClient = &http.Client{
	Transport: transport,
	Timeout:   300 * time.Second,
}

var maxIdleConnsPerHost = beego.AppConfig.DefaultInt("max_idle_conns_per_host", 200)
var transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	ResponseHeaderTimeout: 10 * time.Second,
}

func GetHttpClient(gzip bool) (client *HttpClient, err error) {
	return &HttpClient{
		gzip:   gzip,
		client: httpClient,
	}, nil
}

/**
  获取TLS认证HTTP客户端
  gzip  控制是否支持压缩
  verifyPeer  控制是否认证客户端
  supplyCert  控制是否加载和发送证书
  verifyCN    控制是否认证对端CN
*/
func getHttpsClient(gzip, verifyPeer, supplyCert, verifyCN bool) (client *HttpClient, err error) {
	transport, err := getTLSTransport(verifyPeer, supplyCert, verifyCN)
	if err != nil {
		util.LOGGER.Errorf(err, "get tls transport failed.")
	}

	client = &HttpClient{
		gzip: gzip,
		client: &http.Client{
			Transport: transport,
			Timeout:   300 * time.Second,
		},
	}

	return client, nil
}

/**
  获取TLS认证HTTP客户端（支持压缩，提供证书，是否认证对端通过参数控制）
*/
func GetHttpsClient(verifyPeer bool) (client *HttpClient, err error) {
	return getHttpsClient(true, verifyPeer, true, false)
}

/**
  获取匿名认证HTTP客户端（支持压缩，不提供证书，也不认证对端）
*/
func GetAnnoHttpsClient(gzip bool) (client *HttpClient, err error) {
	return getHttpsClient(gzip, false, false, false)
}

func (client *HttpClient) getHeaders(method string, headers map[string]string, body interface{}) map[string]string {
	newHeaders := make(map[string]string)
	if body != nil {
		newHeaders["Content-Type"] = "application/json;utf-8"
		newHeaders["Accept"] = "application/json"
	}

	if client.gzip {
		newHeaders["Accept-Encoding"] = "gzip"
		newHeaders["Content-Encoding"] = "gzip"
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
		util.LOGGER.Errorf(err, "duplicate gzip reader failed.")
		return nil, err
	}

	defer gzipReader.Close()
	dst, err = ioutil.ReadAll(gzipReader)
	if err != nil {
		util.LOGGER.Errorf(err, "read from gzip reader failed.")
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
		if headers == nil || len(headers["Content-Type"]) == 0 {
			// 如果请求头未传入Conent-Type，则按照json格式进行编码（如果是非json类型，需要自行在headers里指定类型）
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				util.LOGGER.Errorf(err, "mashal object failed.")
				return status, result
			}
		} else {
			// 如果指定了Content-Type类型，则传入的body必须为byte流
			var ok bool = false
			bodyBytes, ok = body.([]byte)
			if !ok {
				util.LOGGER.Errorf(nil, "invalid body type '%s'(%s), body must type of byte array if Content-Type specified.", reflect.TypeOf(body), headers["Content-Type"])
				return status, result
			}
		}

		//如果配置了gzip压缩，则对body压缩一次(如果请求头里传入已经gzip压缩了，则不重复压缩)
		if client.gzip && (headers == nil || headers["Content-Encoding"] != "gzip") {
			bodyBytes = gzipCompress(bodyBytes)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		util.LOGGER.Errorf(err, "create request failed.")
		return status, result
	}

	newHeaders := client.getHeaders(method, headers, body)
	for key, value := range newHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		util.LOGGER.Errorf(err, "invoke request failed.")
		return status, result
	}

	defer resp.Body.Close()
	status = resp.StatusCode
	var respBody []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		// 如果响应头里包含了响应消息的压缩格式为gzip，则在返回前先解压缩
		respBody, _ = readAndGunzip(resp.Body)
	} else {
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	result = string(respBody)

	return status, result
}

func (client *HttpClient) HttpDo(method string, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte = nil
	var err error = nil
	var bodyReader io.Reader = nil
	if body != nil {
		if headers == nil || len(headers["Content-Type"]) == 0 {
			// 如果请求头未传入Conent-Type，则按照json格式进行编码（如果是非json类型，需要自行在headers里指定类型）
			bodyBytes, err = json.Marshal(body)
			if err != nil {
				util.LOGGER.Errorf(err, "mashal object failed.")
				return nil, err
			}
		} else {
			// 如果指定了Content-Type类型，则传入的body必须为byte流
			var ok bool = false
			bodyBytes, ok = body.([]byte)
			if !ok {
				err := errors.New(fmt.Sprintf("invalid body type '%s'(%s), body must type of byte array if Content-Type specified.", reflect.TypeOf(body), headers["Content-Type"]))
				util.LOGGER.Errorf(err, "")
				return nil, err
			}
		}

		//如果配置了gzip压缩，则对body压缩一次(如果请求头里传入已经gzip压缩了，则不重复压缩)
		if client.gzip && (headers == nil || headers["Content-Encoding"] != "gzip") {
			bodyBytes = gzipCompress(bodyBytes)
		}

		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		util.LOGGER.Errorf(err, "create request failed.")
		return nil, err
	}

	newHeaders := client.getHeaders(method, headers, body)
	for key, value := range newHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		util.LOGGER.Errorf(err, "Request -----> %s failed.", url)
		return resp, err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		util.LOGGER.Errorf(nil, "Request -----> %s not ok, status is %s", url, resp.Status)
		return resp, errors.New(fmt.Sprintf("Request failed, status is %s", resp.Status))
	}
	return resp, err
}

func (client *HttpClient) Get(url string, headers map[string]string) (int, string) {
	return client.httpDo(HTTP_METHOD_GET, url, headers, nil)
}

func (client *HttpClient) Put(url string, headers map[string]string, body interface{}) (int, string) {
	return client.httpDo(HTTP_METHOD_PUT, url, headers, body)
}

func (client *HttpClient) Post(url string, headers map[string]string, body interface{}) (int, string) {
	return client.httpDo(HTTP_METHOD_POST, url, headers, body)
}

func (client *HttpClient) Delete(url string, headers map[string]string) (int, string) {
	return client.httpDo(HTTP_METHOD_DELETE, url, headers, nil)
}

func (client *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.client.Do(req)
}
