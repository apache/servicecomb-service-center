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
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/tlsutil"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	"github.com/apache/servicecomb-service-center/syncer/http"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/serf"
	"github.com/apache/servicecomb-service-center/syncer/task"
)

const (
	tagKeyClusterName = "syncer-cluster-name"
	tagKeyClusterPort = "syncer-cluster-port"
	tagKeyRPCPort     = "syncer-rpc-port"
	tagKeyTLSEnabled  = "syncer-tls-enabled"

	groupExpect = 3
)

func convertSerfOptions(c *config.Config) []serf.Option {
	bindHost, bindPort, _ := utils.ResolveAddr(c.Listener.BindAddr)
	_, rpcPort, _ := utils.ResolveAddr(c.Listener.RPCAddr)
	opts := []serf.Option{
		serf.WithNode(c.Node),
		serf.WithBindAddr(bindHost),
		serf.WithBindPort(bindPort),
		serf.WithAddTag(tagKeyRPCPort, strconv.Itoa(rpcPort)),
		serf.WithAddTag(tagKeyTLSEnabled, strconv.FormatBool(c.Listener.TLSMount.Enabled)),
	}

	if c.Cluster != "" {
		_, peerPort, _ := utils.ResolveAddr(c.Listener.PeerAddr)
		opts = append(opts,
			serf.WithAddTag(tagKeyClusterName, c.Cluster),
			serf.WithAddTag(tagKeyClusterPort, strconv.Itoa(peerPort)),
		)
	}
	return opts
}

func convertEtcdOptions(c *config.Config) []etcd.Option {
	return []etcd.Option{
		etcd.WithName(c.Node),
		etcd.WithDataDir(c.DataDir),
		etcd.WithPeerAddr(c.Listener.PeerAddr),
	}
}

func convertGRPCOptions(c *config.Config) []grpc.Option {
	opts := []grpc.Option{
		grpc.WithAddr(c.Listener.RPCAddr),
	}
	if c.Listener.TLSMount.Enabled {
		conf := c.GetTLSConfig(c.Listener.TLSMount.Name)
		sslOps := append(tlsutil.DefaultServerTLSOptions(), tlsConfigToOptions(conf)...)
		tlsConf, err := tlsutil.GetServerTLSConfig(sslOps...)
		if err != nil {

		}
		opts = append(opts, grpc.WithTLSConfig(tlsConf))
	}
	return opts
}

func convertHttpOptions(c *config.Config) []http.Option {
	opts := []http.Option{
		http.WithAddr(c.HttpConfig.HttpAddr),
		http.WithCompressed(c.HttpConfig.Compressed),
		http.WithCompressMinBytes(c.HttpConfig.CompressMinBytes),
	}

	if c.Listener.TLSMount.Enabled {
		conf := c.GetTLSConfig(c.Listener.TLSMount.Name)
		sslOps := append(tlsutil.DefaultServerTLSOptions(), tlsConfigToOptions(conf)...)
		tlsConf, err := tlsutil.GetServerTLSConfig(sslOps...)
		if err != nil {

		}
		opts = append(opts, http.WithTLSConfig(tlsConf))
	}
	return opts
}

func convertTaskOptions(c *config.Config) []task.Option {
	opts := make([]task.Option, 0, len(c.Task.Params))
	for _, label := range c.Task.Params {
		opts = append(opts, task.WithAddKV(label.Key, label.Value))
	}
	return opts
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
