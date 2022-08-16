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
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Mux struct {
	// Disable represents frontend proxy service api or not
	Disable bool
}

func (m *Mux) SchemaHandleFunc(c echo.Context) (err error) {
	if m.Disable {
		c.Response().WriteHeader(http.StatusForbidden)
		_, _ = c.Response().Write([]byte("schema is disabled"))
		return
	}

	r := c.Request()

	//	protocol:= r.Header.Get("X-InstanceProtocol")
	//	sslActive:=r.Header.Get("X-InstanceSSL")
	var (
		response   *http.Response
		req        *http.Request
		instanceIP = r.Header.Get("X-InstanceIP")
		requestURL = strings.Replace(r.RequestURI, "testSchema/", "", 1)
		url        = "http://" + instanceIP + requestURL
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
		oerr := c.String(http.StatusNotFound, "Method not found")
		if oerr != nil {
			log.Printf("Error: %s\n", oerr)
		}
		return

	}

	if err != nil {
		oerr := c.String(http.StatusInternalServerError,
			fmt.Sprintf("( Error while creating request due to : %s", err))
		if oerr != nil {
			log.Printf("Error: %s\n", oerr)
		}
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
		oerr := c.String(http.StatusNotFound,
			fmt.Sprintf("( Error while sending request due to : %s", err))
		if oerr != nil {
			log.Printf("Error: %s\n", oerr)
		}
		return
	}
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		oerr := c.String(http.StatusNotFound,
			fmt.Sprintf("(could not fetch response body for error %s", err))
		if oerr != nil {
			log.Printf("Error: %s\n", oerr)
		}
		return
	}

	oerr := c.String(http.StatusOK, string(respBody))
	if oerr != nil {
		log.Printf("Error: %s\n", oerr)
	}
	return nil
}
