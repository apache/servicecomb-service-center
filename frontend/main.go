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
	"flag"
	"fmt"

	"net"
	"net/url"
	"strconv"

	"github.com/astaxie/beego"
)

type Config struct {
	frontendAddr string
	scAddr       string
}

func main() {
	frontendIp := beego.AppConfig.String("frontend_host_ip")
	frontendPort := beego.AppConfig.DefaultInt("frontend_host_port", 30103)

	scIp := beego.AppConfig.DefaultString("httpaddr", "127.0.0.1")
	scPort := beego.AppConfig.DefaultInt("httpport", 30100)

	// command line flags
	port := flag.Int("port", frontendPort, "port to serve on")
	flag.Parse()

	cfg := Config{}
	cfg.scAddr = fmt.Sprintf("http://%s/", net.JoinHostPort(url.PathEscape(scIp), strconv.Itoa(scPort)))
	cfg.frontendAddr = net.JoinHostPort(frontendIp, strconv.Itoa(*port))

	// run frontend web server
	Serve(cfg)
}
