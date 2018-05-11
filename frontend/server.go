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
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/apache/incubator-servicecomb-service-center/frontend/schema"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Serve(c Config) {
	e := echo.New()
	e.HideBanner = true
	// handle all requests by serving a file of the same name
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cant get cwd, error:%s", err)
	}
	staticPath := filepath.Join(dir, "app")
	e.Static("/", staticPath)

	e.Any("/testSchema/*", schema.SchemaHandleFunc)

	scProxy(c, e)

	log.Printf("Error: %s", e.Start(c.frontendAddr))
}

// setup proxy for requests to service center
func scProxy(c Config, e *echo.Echo) {
	scUrl, err := url.Parse(c.scAddr)
	if err != nil {
		log.Fatalf("Error parsing service center address:%s, err:%s", c.scAddr, err)
	}

	targets := []*middleware.ProxyTarget{
		{
			URL: scUrl,
		},
	}
	g := e.Group("/sc")
	balancer := middleware.NewRoundRobinBalancer(targets)
	pcfg := middleware.ProxyConfig{
		Balancer: balancer,
		Skipper:  middleware.DefaultSkipper,
		Rewrite: map[string]string{
			"/sc/*": "/$1",
		},
	}
	g.Use(middleware.ProxyWithConfig(pcfg))
}
