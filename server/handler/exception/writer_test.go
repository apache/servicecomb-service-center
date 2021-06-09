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

package exception_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/server/handler/exception"
	"github.com/stretchr/testify/assert"
)

type handler struct {
	Writer *exception.Writer
}

func (receiver *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	aw := exception.NewWriter(w)
	aw.Write([]byte("Bye"))
	aw.Header().Set("a", "b")
	aw.WriteHeader(http.StatusCreated)
	aw.Flush()
}

func TestAsyncWriter_Flush(t *testing.T) {
	h := &handler{}
	c := &http.Client{}
	s := httptest.NewServer(h)
	defer s.Close()

	t.Run("should recv the response when new writer", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, s.URL, bytes.NewReader([]byte("Hi")))
		assert.NoError(t, err)
		resp, err := c.Do(request)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "b", resp.Header.Get("a"))
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Bye", string(body))
	})
}
