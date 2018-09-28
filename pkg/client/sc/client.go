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

package sc

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"io/ioutil"
	"strings"
	"time"
)

func NewSCClient(cfg Config) (*SCClient, error) {
	ssl := strings.Index(cfg.Addr, "https://") >= 0
	if ssl && len(cfg.CertKeyPWD) == 0 && len(cfg.CertKeyPWDPath) > 0 {
		content, _ := ioutil.ReadFile(cfg.CertKeyPWDPath)
		cfg.CertKeyPWD = string(content)
	}
	client, err := rest.GetURLClient(rest.URLClientOption{
		SSLEnabled:     ssl,
		VerifyPeer:     cfg.VerifyPeer,
		CAFile:         cfg.CAFile,
		CertFile:       cfg.CertFile,
		CertKeyFile:    cfg.CertKeyFile,
		CertKeyPWD:     cfg.CertKeyPWD,
		RequestTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &SCClient{Config: cfg, URLClient: client}, nil
}

type SCClient struct {
	*rest.URLClient
	Config Config
}
