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

package etcd

import (
	"net/url"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/coreos/etcd/compactor"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver"
)

type config struct {
	name     string
	dataDir  string
	peerAddr string
	peers    map[string]string
}

// Option to etcd config
type Option func(*config)

// WithName returns name option
func WithName(name string) Option {
	return func(c *config) { c.name = name }
}

// WithDataDir returns data dir option
func WithDataDir(dir string) Option {
	return func(c *config) { c.dataDir = dir }
}

// WithPeerAddr returns peer address option
func WithPeerAddr(peerAddr string) Option {
	return func(c *config) { c.peerAddr = peerAddr }
}

// WithAddPeers returns add peers option
func WithAddPeers(name, addr string) Option {
	return func(c *config) {
		c.peers[name] = addr
	}
}

func toConfig(ops ...Option) *config {
	c := &config{peers: map[string]string{}}
	for _, op := range ops {
		op(c)
	}
	return c
}

func toEtcdConfig(ops ...Option) (*embed.Config, error) {
	conf := embed.NewConfig()
	conf.EnableV2 = false
	conf.EnablePprof = false
	conf.QuotaBackendBytes = etcdserver.MaxQuotaBytes
	conf.AutoCompactionMode = compactor.ModePeriodic
	conf.AutoCompactionRetention = "1h"

	conf.ACUrls = nil
	conf.LCUrls = nil

	err := mergeConfig(conf, toConfig(ops...))
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func mergeConfig(src *embed.Config, dst *config) error {
	if dst.name != "" {
		src.Name = dst.name
	}

	if dst.dataDir != "" {
		src.Dir = dst.dataDir
	}

	proto := "http://"
	if dst.peerAddr != "" {
		peer, err := url.Parse(proto + dst.peerAddr)
		if err != nil {
			log.Errorf(err, "parse peer listener '%s' failed", dst.peerAddr)
			return err
		}
		src.APUrls = []url.URL{*peer}
		src.LPUrls = []url.URL{*peer}
		src.InitialCluster = src.Name + "=" + peer.String()
	}

	if len(dst.peers) > 0 {
		initialCluster := ""
		for key, val := range dst.peers {
			initialCluster += key + "=" + proto + val + ","
		}
		src.InitialCluster = initialCluster[:len(initialCluster)-1]
	}
	return nil
}
