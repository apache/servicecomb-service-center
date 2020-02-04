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
	"crypto/tls"
	"net/url"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/serf"
)

func convertSerfConfig(c *config.Config) *serf.Config {
	conf := serf.DefaultConfig()
	conf.NodeName = c.Node
	conf.ClusterName = c.Cluster
	conf.Mode = c.Mode
	conf.TLSEnabled = c.Listener.TLSMount.Enabled
	conf.BindAddr = c.Listener.BindAddr
	_, conf.ClusterPort, _ = utils.ResolveAddr(c.Listener.PeerAddr)
	_, conf.RPCPort, _ = utils.ResolveAddr(c.Listener.RPCAddr)
	if c.Join.Enabled {
		conf.RetryJoin = strings.Split(c.Join.Address, ",")
		conf.RetryInterval, _ = time.ParseDuration(c.Join.RetryInterval)
		conf.RetryMaxAttempts = c.Join.RetryMax
	}
	return conf
}

func convertEtcdConfig(c *config.Config) *etcd.Config {
	conf := etcd.DefaultConfig()
	conf.Name = c.Node
	conf.Dir = c.DataDir
	proto := "http://"

	if c.Listener.TLSMount.Enabled {
		proto = "https://"
	}

	peer, _ := url.Parse(proto + c.Listener.PeerAddr)
	conf.APUrls = []url.URL{*peer}
	conf.LPUrls = []url.URL{*peer}
	return conf
}

func convertTickerInterval(c *config.Config) int {
	strNum := ""
	for _, label := range c.Task.Params {
		if label.Key == "interval" {
			strNum = label.Value
			break
		}
	}
	interval, _ := time.ParseDuration(strNum)
	return int(interval.Seconds())
}

func convertSCConfigOption(c *config.Config) []plugins.SCConfigOption {
	endpoints := make([]string, 0, 10)
	for _, endpoint := range strings.Split(c.Registry.Address, ",") {
		endpoints = append(endpoints, endpoint)
	}
	opts := []plugins.SCConfigOption{plugins.WithEndpoints(endpoints)}

	if c.Registry.TLSMount.Enabled {
		tlsConf := c.GetTLSConfig(c.Registry.TLSMount.Name)
		opts = append(
			opts, plugins.WithTLSEnabled(c.Registry.TLSMount.Enabled),
			plugins.WithTLSVerifyPeer(tlsConf.VerifyPeer),
			plugins.WithTLSPassphrase(tlsConf.Passphrase),
			plugins.WithTLSCAFile(tlsConf.CAFile),
			plugins.WithTLSCertFile(tlsConf.CertFile),
			plugins.WithTLSKeyFile(tlsConf.KeyFile),
		)
	}
	return opts
}

func tlsConfigToOptions(t *config.TLSConfig) []tlsutil.SSLConfigOption {
	return []tlsutil.SSLConfigOption{
		tlsutil.WithVerifyPeer(t.VerifyPeer),
		tlsutil.WithVersion(tlsutil.ParseSSLProtocol(t.MinVersion), tls.VersionTLS12),
		tlsutil.WithCipherSuits(
			tlsutil.ParseDefaultSSLCipherSuites(strings.Join(t.Ciphers, ","))),
		tlsutil.WithKeyPass(t.Passphrase),
		tlsutil.WithCA(t.CAFile),
		tlsutil.WithCert(t.CertFile),
		tlsutil.WithKey(t.KeyFile),
	}
}
