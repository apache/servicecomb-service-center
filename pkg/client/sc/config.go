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
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"io/ioutil"
	"strings"
	"time"
)

const defaultRequestTimeout = 10 * time.Second

type Config struct {
	rest.URLClientOption
	Name      string
	Endpoints []string
	// TODO Expandable header not only token header
	Token          string
	CertKeyPWDPath string
}

func (cfg *Config) Merge() rest.URLClientOption {
	ssl := strings.Contains(cfg.Endpoints[0], "https://")
	if ssl && len(cfg.CertKeyPWD) == 0 && len(cfg.CertKeyPWDPath) > 0 {
		content, _ := ioutil.ReadFile(cfg.CertKeyPWDPath)
		cfg.CertKeyPWD = string(content)
	}
	cfg.SSLEnabled = ssl
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = defaultRequestTimeout
	}
	cfg.Compressed = true
	return cfg.URLClientOption
}
