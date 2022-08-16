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

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/rest"
)

func TestNewLBClient(t *testing.T) {
	os.Setenv("DEBUG_MODE", "1")
	client, err := NewLBClient([]string{"x.x.x.x", "rest://2.2.2.2"}, rest.DefaultURLClientOption())
	if err != nil {
		t.Fatal("TestNewLBClient", err)
	}
	_, err = client.RestDoWithContext(context.Background(), "yyy", "/zzz", http.Header{"test": []string{"a"}}, []byte(`abcdef`))
	if err == nil {
		t.Fatal("TestNewLBClient")
	}
	fmt.Println(err)

	svc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, _ := io.ReadAll(req.Body)
		w.Write(b)
	}))
	defer svc.Close()

	client, err = NewLBClient([]string{"x.x.x.x", svc.URL}, rest.DefaultURLClientOption())
	if err != nil {
		t.Fatal("TestNewLBClient", err)
	}
	_, err = client.RestDoWithContext(context.Background(), http.MethodGet, "", http.Header{"test": []string{"a"}}, []byte(`abcdef`))
	if err != nil {
		t.Fatal("TestNewLBClient", err)
	}
}
