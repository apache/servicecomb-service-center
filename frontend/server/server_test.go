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
package server

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

const (
	SCAddr    = "127.0.0.1:30101"
	FrontAddr = "127.0.0.1:30104"
)

func init() {
	wd, _ := os.Getwd()
	_ = os.Chdir(filepath.Join(wd, "../"))
}

func TestStatic(t *testing.T) {
	cfg := Config{
		SCAddr:       "http://" + SCAddr,
		FrontendAddr: FrontAddr,
	}

	go Serve(cfg)
	time.Sleep(500 * time.Millisecond)

	res, err := http.Get("http://" + FrontAddr)
	assert.NoError(t, err, "Error accessing frontend: %s", err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "Expected http %d, got %d", http.StatusOK, res.StatusCode)
	_ = res.Body.Close()
}

func TestSCProxy(t *testing.T) {
	greeting := "Hi, there!"

	// simulate service center backend
	go func() {
		e := echo.New()
		e.HideBanner = true
		e.GET("/sayHi", func(c echo.Context) error {
			return c.String(http.StatusOK, greeting)
		})
		_ = e.Start(SCAddr)
	}()
	time.Sleep(500 * time.Millisecond)

	res, err := http.Get("http://" + FrontAddr + "/sc/sayHi")
	assert.NoError(t, err, "Error accessing sc proxy: %s", err)
	assert.Equal(t, http.StatusOK, res.StatusCode, "Expected http %d, got %d", http.StatusOK, res.StatusCode)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error reading body: %s", err)
	}
	if string(body) != greeting {
		t.Errorf("Expected %s, got %s", greeting, string(body))
	}
}

func TestDirectoryTraversal(t *testing.T) {
	cfg := Config{
		SCAddr:       "http://" + SCAddr,
		FrontendAddr: FrontAddr,
	}

	go Serve(cfg)
	time.Sleep(500 * time.Millisecond)

	res, err := http.Get("http://" + FrontAddr + "/..\\schema/schemahandler.go")
	assert.NoError(t, err, "Error accessing frontend: %s", err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode, "Expected http status is 404")
	_ = res.Body.Close()
}
