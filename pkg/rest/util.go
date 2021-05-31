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
	"encoding/json"
	"errors"
	"io/ioutil"
	"github.com/go-chassis/cari/pkg/errsvc"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/cari/discovery"
)

var errNilRequestBody = errors.New("request body is nil")

func WriteError(w http.ResponseWriter, code int32, detail string) {
	err := discovery.NewError(code, detail)
	WriteErrsvcError(w, err)
}

func WriteErrsvcError(w http.ResponseWriter, err *errsvc.Error) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(err.StatusCode())
	b, _ := json.Marshal(err)
	_, _ = w.Write(b)
}

// WriteResponse writes http response
// If the resp is nil or represents success, response status is http.StatusOK,
// response content is obj.
// If the resp represents fail, response status is from the code in the
// resp, response content is from the message in the resp.
func WriteResponse(w http.ResponseWriter, r *http.Request, resp *discovery.Response, obj interface{}) {
	if resp != nil && resp.GetCode() != discovery.ResponseSuccess {
		WriteError(w, resp.GetCode(), resp.GetMessage())
		return
	}

	if obj == nil {
		w.Header().Set(HeaderContentType, ContentTypeText)
		w.WriteHeader(http.StatusOK)
		return
	}

	util.SetRequestContext(r, CtxResponseObject, obj)

	var (
		data []byte
		err  error
	)
	switch body := obj.(type) {
	case []byte:
		data = body
	default:
		data, err = json.Marshal(body)
		if err != nil {
			WriteError(w, discovery.ErrInternal, err.Error())
			return
		}
	}
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		log.Error("write response failed", err)
	}
}

func WriteSuccess(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, r, nil, nil)
}

// ReadBody can re-read the request body
func ReadBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, errNilRequestBody
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(data))
	return data, nil
}
