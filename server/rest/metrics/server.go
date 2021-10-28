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

package metrics

import (
	"context"
	"net/http"
	"net/url"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	restsvc "github.com/apache/servicecomb-service-center/server/rest"
	"github.com/go-chassis/foundation/gopool"
)

const (
	DefaultAPIPath = "/metrics"
)

func ListenAndServe() error {
	cfg, listenURL, err := getConfig()
	if err != nil {
		return err
	}
	if listenURL == "" {
		restsvc.RegisterServerHandler(DefaultAPIPath, prometheus.HTTPHandler())
		return nil
	}

	srv := rest.NewServer(cfg)
	if cfg.TLSConfig == nil {
		err = srv.Listen()
	} else {
		err = srv.ListenTLS()
	}
	if err != nil {
		return err
	}
	log.Info("listen metrics address: " + listenURL)

	gopool.Go(func(_ context.Context) {
		if err := srv.Serve(); err != nil {
			log.Error("metrics server serve failed", err)
			return
		}
	})
	return nil
}

func getConfig() (*rest.ServerConfig, string, error) {
	listenURL := config.GetString("metrics.prometheus.listenURL", "")
	if listenURL == "" {
		return nil, "", nil
	}
	u, err := url.Parse(listenURL)
	if err != nil {
		log.Error("parse metrics server listenURL failed", err)
		return nil, "", err
	}
	apiPath := DefaultAPIPath
	if u.Path != "" {
		apiPath = u.Path
	}
	mux := http.NewServeMux()
	mux.Handle(apiPath, prometheus.HTTPHandler())

	cfg := rest.DefaultServerConfig()
	cfg.Addr = u.Host
	cfg.Handler = mux
	if u.Scheme != "http" {
		cfg.TLSConfig, err = tlsconf.ServerConfig()
		if err != nil {
			log.Error("get metrics server tls config failed", err)
			return nil, "", err
		}
	}
	return cfg, listenURL, nil
}
