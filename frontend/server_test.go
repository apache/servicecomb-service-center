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
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/labstack/echo"
)

const (
	SCAddr    = "127.0.0.1:30101"
	FrontAddr = "127.0.0.1:30104"
)

func TestStatic(t *testing.T) {
	var wg sync.WaitGroup

	cfg := Config{
		scAddr:       "http://" + SCAddr,
		frontendAddr: FrontAddr,
	}

	wg.Add(1)
	go func() {
		wg.Done()
		Serve(cfg)
	}()

	wg.Wait()
	res, err := http.Get("http://" + FrontAddr)
	if err != nil {
		t.Errorf("Error accessing frontend: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected http %d, got %d", http.StatusOK, res.StatusCode)
	}

}

func TestSCProxy(t *testing.T) {
	var wg sync.WaitGroup
	greeting := "Hi, there!"

	wg.Add(1)
	// simulate service center backend
	go func() {
		e := echo.New()
		e.HideBanner = true
		e.GET("/sayHi", func(c echo.Context) error {
			return c.String(http.StatusOK, greeting)
		})
		wg.Done()
		e.Start(SCAddr)
	}()

	wg.Wait()
	res, err := http.Get("http://" + FrontAddr + "/sc/sayHi")
	if err != nil {
		t.Errorf("Error accessing sc proxy: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected http %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error reading body: %s", err)
	}
	if string(body) != greeting {
		t.Errorf("Expected %s, got %s", greeting, string(body))
	}

}
