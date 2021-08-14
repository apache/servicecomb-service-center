// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"crypto/tls"
	"io/ioutil"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/go-chassis/foundation/tlsutil"
)

func NewEtcdClient(cfg Config) (*clientv3.Client, error) {
	var (
		endpoints = strings.Split(cfg.Addrs, ",")
		cliTls    *tls.Config
	)
	for _, ip := range endpoints {
		if strings.Contains(ip, "https://") {
			if len(cfg.CertKeyPWD) == 0 && len(cfg.CertKeyPWDPath) > 0 {
				content, _ := ioutil.ReadFile(cfg.CertKeyPWDPath)
				cfg.CertKeyPWD = string(content)
			}
			opts := append(tlsutil.DefaultClientTLSOptions(),
				tlsutil.WithCA(cfg.CAFile),
				tlsutil.WithCert(cfg.CertFile),
				tlsutil.WithKey(cfg.CertKeyFile),
				tlsutil.WithKeyPass(cfg.CertKeyPWD))
			cliTls, _ = tlsutil.GetClientTLSConfig(opts...)
			break
		}
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 10 * time.Second,
		TLS:         cliTls,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
