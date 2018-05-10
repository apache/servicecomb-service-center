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

	"github.com/labstack/echo"
)

func SchemaHandleFunc(c echo.Context) (err error) {
	r := c.Request()

	//	protocol:= r.Header.Get("X-InstanceProtocol")
	//	sslActive:=r.Header.Get("X-InstanceSSL")
	var (
		response   *http.Response
		req        *http.Request
		instanceIP = r.Header.Get("X-InstanceIP")
		requestUrl = strings.Replace(r.RequestURI, "testSchema/", "", -1)
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
		c.String(http.StatusNotFound, "Method not found")
		return

	}

	if err != nil {
		c.String(http.StatusInternalServerError,
			fmt.Sprintf("( Error while creating request due to : %s", err))
		return
	}

	for key, values := range r.Header {
		for _, val := range values {
			if key == "Accept-Encoding" || key == "Connection" || key == "X-Schemaname" || key == "Cookie" || key == "User-Agent" || key == "AppleWebKit" || key == "Dnt" || key == "Referer" || key == "Accept-Language" {
				continue
			} else {
				req.Header.Add(key, val)
			}

		}
	}
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{Timeout: time.Second * 20}
	response, err = client.Do(req)
	if err != nil {
		c.String(http.StatusNotFound,
			fmt.Sprintf("( Error while sending request due to : %s", err))
		return
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.String(http.StatusNotFound,
			fmt.Sprintf("(could not fetch response body for error %s", err))
		return
	}

	c.String(http.StatusOK, string(respBody))
	return nil
}
