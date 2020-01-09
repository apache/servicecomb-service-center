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
package interceptor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var i = 0

func mockFunc(w http.ResponseWriter, r *http.Request) error {
	switch i {
	case 1:
		return errors.New("error")
	case 0:
		panic(errors.New("panic"))
	default:
		i++
	}
	return nil
}

func TestInterception_Invoke(t *testing.T) {
	RegisterInterceptFunc(mockFunc)
	RegisterInterceptFunc(mockFunc)
	RegisterInterceptFunc(mockFunc)

	i = 1
	err := InvokeInterceptors(nil, nil)
	if err == nil || err.Error() != "error" {
		t.Fatalf("TestInterception_Invoke failed")
	}

	i = 0
	err = InvokeInterceptors(httptest.NewRecorder(), nil)
	if err == nil || err.Error() != "panic" {
		t.Fatalf("TestInterception_Invoke failed")
	}

	i = 2
	err = InvokeInterceptors(nil, nil)
	if err != nil || i != 2+len(interceptors) {
		t.Fatalf("TestInterception_Invoke failed")
	}
}
