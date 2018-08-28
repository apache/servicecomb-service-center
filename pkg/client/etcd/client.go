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
	"github.com/apache/incubator-servicecomb-service-center/pkg/tlsutil"
	"github.com/coreos/etcd/clientv3"
	"io/ioutil"
	"strings"
	"time"
)

var (
	Addrs       string
	CertPath    string
	KeyPath     string
	KeyPassPath string
	KeyPass     string
	CAPath      string
)

func NewEtcdClient() (*clientv3.Client, error) {
	var (
		endpoints = strings.Split(Addrs, ",")
		cliTls    *tls.Config
	)
	for _, ip := range endpoints {
		if strings.Index(ip, "https://") >= 0 {
			if len(KeyPass) == 0 && len(KeyPassPath) > 0 {
				content, _ := ioutil.ReadFile(KeyPassPath)
				KeyPass = string(content)
			}
			opts := append(tlsutil.DefaultClientTLSOptions(),
				tlsutil.WithCA(CAPath),
				tlsutil.WithCert(CertPath),
				tlsutil.WithKey(KeyPath),
				tlsutil.WithKeyPass(KeyPass))
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
