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

package service

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) HandleSqrt(w http.ResponseWriter, r *http.Request) {
	value := 0
	var err error
	keyValue, ok := r.URL.Query()["x"]
	if ok {
		value, err = strconv.Atoi(keyValue[0])
		if err != nil {
			fmt.Printf("input value error: %s", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	fmt.Println("handle request")
	resp, err := http.Get(fmt.Sprintf("http://provider:8081/sqrt?x=%d", value))
	if err != nil {
		fmt.Printf("http get provider error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("read http body error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, fmt.Sprintf("Get result from microservice provider: %s", string(body)))
}
