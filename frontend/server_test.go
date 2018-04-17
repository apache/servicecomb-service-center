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
package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo"
)

const (
	SCAddr    = "127.0.0.1:30101"
	FrontAddr = "127.0.0.1:30104"
)

func TestStatic(t *testing.T) {
	cfg := Config{
		scAddr:       "http://" + SCAddr,
		frontendAddr: FrontAddr,
	}

	go Serve(cfg)
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://" + FrontAddr)
	if err != nil {
		t.Errorf("Error accessing frontend: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected http %d, got %d", http.StatusOK, res.StatusCode)
	}

}

func TestSCProxy(t *testing.T) {
	// simulate service center backend
	go func() {
		e := echo.New()
		e.HideBanner = true
		e.GET("/sayHi", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hi, there!")
		})
		e.Start(SCAddr)
	}()
	time.Sleep(1 * time.Second)

	res, err := http.Get("http://" + FrontAddr + "/sc/sayHi")
	if err != nil {
		t.Errorf("Error accessing sc proxy: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected http %d, got %d", http.StatusOK, res.StatusCode)
	}

}
