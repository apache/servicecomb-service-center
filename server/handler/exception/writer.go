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

package exception

import (
	"fmt"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

// Writer is the async response writer, it is not thread safe!
type Writer struct {
	StatusCode int
	Body       []byte
	w          http.ResponseWriter
	flushed    bool
}

func (aw *Writer) Header() http.Header {
	return aw.w.Header()
}

func (aw *Writer) Write(body []byte) (int, error) {
	aw.Body = body
	return len(body), nil
}

func (aw *Writer) WriteHeader(statusCode int) {
	aw.StatusCode = statusCode
}

func (aw *Writer) Flush() error {
	if aw.flushed {
		log.Warn("writer already flushed")
		return nil
	}
	aw.flushed = true

	if aw.StatusCode == 0 {
		err := fmt.Errorf("unknown status code %d", aw.StatusCode)
		aw.w.Header().Set(rest.HeaderContentType, rest.ContentTypeText)
		aw.w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := aw.w.Write(util.StringToBytesWithNoCopy(err.Error())); writeErr != nil {
			log.Error("write response failed", writeErr)
		}
		return err
	}
	aw.w.WriteHeader(aw.StatusCode)
	if len(aw.Body) == 0 {
		return nil
	}

	_, err := aw.w.Write(aw.Body)
	return err
}

func NewWriter(w http.ResponseWriter) *Writer {
	return &Writer{w: w}
}
