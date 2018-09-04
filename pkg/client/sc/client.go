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

var (
	Addr        string
	Token       string
	VerifyPeer  bool
	CertPath    string
	KeyPath     string
	KeyPassPath string
	KeyPass     string
	CAPath      string
)

func NewSCClient() (*rest.URLClient, error) {
	ssl := strings.Index(Addr, "https://") >= 0
	if ssl && len(KeyPass) == 0 && len(KeyPassPath) > 0 {
		content, _ := ioutil.ReadFile(KeyPassPath)
		KeyPass = string(content)
	}
	return rest.GetURLClient(&rest.URLClientOption{
		SSLEnabled:     ssl,
		VerifyPeer:     VerifyPeer,
		CAFile:         CAPath,
		CertFile:       CertPath,
		CertKeyFile:    KeyPath,
		CertKeyPWD:     KeyPass,
		RequestTimeout: 10 * time.Second,
	})
}
