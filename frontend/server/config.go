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
	"flag"
	"fmt"
	"net"
	"net/url"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
)

type Config struct {
	FrontendAddr string
	SCAddr       string
}

func DefaultConfig() Config {
	frontendIP := beego.AppConfig.DefaultString("frontend_host_ip", "127.0.0.1")
	frontendPort := beego.AppConfig.DefaultInt("frontend_host_port", 30103)

	scIP := beego.AppConfig.DefaultString("httpaddr", "127.0.0.1")
	scPort := beego.AppConfig.DefaultInt("httpport", 30100)

	// command line flags
	port := flag.Int("port", frontendPort, "port to serve on")
	flag.Parse()

	cfg := Config{}
	cfg.SCAddr = fmt.Sprintf("http://%s/", net.JoinHostPort(url.PathEscape(scIP), strconv.Itoa(scPort)))
	cfg.FrontendAddr = net.JoinHostPort(frontendIP, strconv.Itoa(*port))
	return cfg
}
