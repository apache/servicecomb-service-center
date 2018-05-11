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
package error

import (
	"net/http"
	"testing"
)

func TestError_StatusCode(t *testing.T) {
	e := Error{Code: 503999}
	if e.StatusCode() != http.StatusServiceUnavailable {
		t.Fatalf("TestError_StatusCode %v failed", e)
	}

	if !e.InternalError() {
		t.Fatalf("TestInternalError failed")
	}
}

func TestNewError(t *testing.T) {
	var err error
	err = NewError(ErrInvalidParams, "test1")
	if err == nil {
		t.Fatalf("TestNewError failed")
	}
	err = NewErrorf(ErrInvalidParams, "%s", "test2")
	if err == nil {
		t.Fatalf("TestNewErrorf failed")
	}

	if len(err.Error()) == 0 {
		t.Fatalf("TestError failed")
	}

	if len(err.(*Error).Marshal()) == 0 {
		t.Fatalf("TestMarshal failed")
	}

	if err.(*Error).StatusCode() != http.StatusBadRequest {
		t.Fatalf("TestStatusCode failed, %d", err.(*Error).StatusCode())
	}

	if err.(*Error).InternalError() {
		t.Fatalf("TestInternalError failed")
	}

	err = NewErrorf(ErrInvalidParams, "")
	if len(err.Error()) == 0 {
		t.Fatalf("TestNewErrorf with empty detial failed")
	}
}

func TestRegisterErrors(t *testing.T) {
	RegisterErrors(map[int32]string{503999: "test1", 1: "none"})

	e := NewError(503999, "test2")
	if e.Message != "test1" {
		t.Fatalf("TestRegisterErrors failed")
	}
}
