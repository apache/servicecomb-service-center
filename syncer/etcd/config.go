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
	"fmt"
	"net/url"
	"os"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/coreos/etcd/compactor"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver"
)

var (
	defaultDataDir        = "syncer-data/"
	defaultListenPeerAddr = "http://127.0.0.1:39102"
)

type Config struct {
	*embed.Config
}

// DefaultConfig default configure of etcd
func DefaultConfig() *Config {
	etcdConf := embed.NewConfig()
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf(err, "Error determining hostname: %s", err)
		return nil
	}

	peer, _ := url.Parse(defaultListenPeerAddr)
	etcdConf.ACUrls = nil
	etcdConf.LCUrls = nil
	etcdConf.APUrls = []url.URL{*peer}
	etcdConf.LPUrls = []url.URL{*peer}

	etcdConf.EnableV2 = false
	etcdConf.EnablePprof = false
	etcdConf.QuotaBackendBytes = etcdserver.MaxQuotaBytes
	etcdConf.Dir = defaultDataDir + hostname
	etcdConf.Name = hostname
	etcdConf.InitialCluster = fmt.Sprintf("%s=%s", hostname, defaultListenPeerAddr)
	etcdConf.AutoCompactionMode = compactor.ModePeriodic
	etcdConf.AutoCompactionRetention = "1h"
	return &Config{Config: etcdConf}
}
