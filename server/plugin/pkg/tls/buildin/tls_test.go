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
package buildin

import (
	"crypto/tls"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/pkg/security/buildin"
	"os"
	"testing"
)

func init() {
	core.Initialize()

	sslRoot := "../../../../../examples/service_center/ssl/"
	os.Setenv("SSL_ROOT", sslRoot)
}

func TestGetServerTLSConfig(t *testing.T) {
	serverTLSConfig, err := GetServerTLSConfig()
	if err != nil {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if len(serverTLSConfig.Certificates) == 0 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.ClientCAs == nil {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if len(serverTLSConfig.CipherSuites) != 4 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.MaxVersion != tls.VersionTLS12 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Fatalf("GetServerTLSConfig failed")
	}
}

func TestGetClientTLSConfig(t *testing.T) {
	clientTLSConfig, err := GetClientTLSConfig()
	if err != nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if len(clientTLSConfig.Certificates) == 0 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.RootCAs == nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if len(clientTLSConfig.CipherSuites) != 0 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.MaxVersion != tls.VersionTLS12 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.InsecureSkipVerify != true {
		t.Fatalf("GetClientTLSConfig failed")
	}
}
