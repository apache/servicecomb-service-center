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

package rest

import (
	"crypto/tls"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"time"
)

func LoadConfig() (srvCfg *rest.ServerConfig, err error) {
	srvCfg = rest.DefaultServerConfig()
	readHeaderTimeout, _ := time.ParseDuration(core.ServerInfo.Config.ReadHeaderTimeout)
	readTimeout, _ := time.ParseDuration(core.ServerInfo.Config.ReadTimeout)
	idleTimeout, _ := time.ParseDuration(core.ServerInfo.Config.IdleTimeout)
	writeTimeout, _ := time.ParseDuration(core.ServerInfo.Config.WriteTimeout)
	maxHeaderBytes := int(core.ServerInfo.Config.MaxHeaderBytes)
	var tlsConfig *tls.Config
	if core.ServerInfo.Config.SslEnabled {
		tlsConfig, err = plugin.Plugins().TLS().ServerConfig()
		if err != nil {
			return
		}
	}
	srvCfg.ReadHeaderTimeout = readHeaderTimeout
	srvCfg.ReadTimeout = readTimeout
	srvCfg.IdleTimeout = idleTimeout
	srvCfg.WriteTimeout = writeTimeout
	srvCfg.MaxHeaderBytes = maxHeaderBytes
	srvCfg.TLSConfig = tlsConfig
	srvCfg.Handler = DefaultServerMux
	return
}

func NewServer(ipAddr string) (srv *rest.Server, err error) {
	srvCfg, err := LoadConfig()
	if err != nil {
		return
	}
	srvCfg.Addr = ipAddr
	srv = rest.NewServer(srvCfg)

	if srvCfg.TLSConfig == nil {
		err = srv.Listen()
	} else {
		err = srv.ListenTLS()
	}
	if err != nil {
		return
	}
	return
}
