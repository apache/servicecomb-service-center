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

package schema

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type schemaHandler struct {
}

func TestSchema() http.Handler {
	return &schemaHandler{}
}

func (sc schemaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//	protocol:= r.Header.Get("X-InstanceProtocol")
	//	sslActive:=r.Header.Get("X-InstanceSSL")
	var (
		response   *http.Response
		err        error
		req        *http.Request
		instanceIP = r.Header.Get("X-InstanceIP")
		schemaName = r.Header.Get("X-SchemaName")
		requestUrl = strings.Replace(r.RequestURI, "testSchema", schemaName, -1)
		url        = "http://" + instanceIP + requestUrl
	)

	switch r.Method {
	case "GET":
		req, err = http.NewRequest(http.MethodGet, url, nil)
	case "POST":
		req, err = http.NewRequest(http.MethodPost, url, r.Body)
	case "PUT":
		req, err = http.NewRequest(http.MethodPut, url, r.Body)
	case "DELETE":
		req, err = http.NewRequest(http.MethodDelete, url, r.Body)
	default:
		w.Write([]byte("Method not found"))
		return

	}

	if err != nil {
		w.Write([]byte(fmt.Sprintf("( Error while creating request due to : %s", err)))
		return
	}

	for key, values := range r.Header {
		for _, val := range values {
			req.Header.Add(key, val)
		}
	}

	client := http.Client{Timeout: time.Second * 20}
	response, err = client.Do(req)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("( Error while sending request due to : %s", err)))
		return
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("(could not fetch response body for error %s", err)))
		return
	}

	w.Write(respBody)
}
