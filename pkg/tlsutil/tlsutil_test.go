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

package tlsutil

import (
	"crypto/tls"
	"io/ioutil"
	"testing"
)

const sslRoot = "../../examples/service_center/ssl/"

func TestParseDefaultSSLCipherSuites(t *testing.T) {
	c := ParseDefaultSSLCipherSuites("")
	if c != nil {
		t.Fatalf("ParseDefaultSSLCipherSuites failed")
	}
	c = ParseDefaultSSLCipherSuites("TLS_RSA_WITH_AES_128_CBC_SHA256")
	if len(c) != 1 {
		t.Fatalf("ParseDefaultSSLCipherSuites failed")
	}
	c = ParseDefaultSSLCipherSuites("a")
	if len(c) != 0 {
		t.Fatalf("ParseDefaultSSLCipherSuites failed")
	}
	c = ParseDefaultSSLCipherSuites("a,,b")
	if len(c) != 0 {
		t.Fatalf("ParseDefaultSSLCipherSuites failed")
	}
}

func TestGetServerTLSConfig(t *testing.T) {
	pw, _ := ioutil.ReadFile(sslRoot + "cert_pwd")
	opts := append(DefaultServerTLSOptions(),
		WithVerifyPeer(true),
		WithVersion(ParseSSLProtocol("TLSv1.0"), tls.VersionTLS12),
		WithCipherSuits(ParseDefaultSSLCipherSuites("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")),
		WithKeyPass(string(pw)),
		WithCA(sslRoot+"trust.cer"),
		WithCert(sslRoot+"server.cer"),
		WithKey(sslRoot+"server_key.pem"),
	)
	serverTLSConfig, err := GetServerTLSConfig(opts...)
	if err != nil {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if len(serverTLSConfig.Certificates) == 0 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.ClientCAs == nil {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if len(serverTLSConfig.CipherSuites) != 2 {
		t.Fatalf("GetServerTLSConfig failed")
	}
	if serverTLSConfig.MinVersion != tls.VersionTLS10 {
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
	pw, _ := ioutil.ReadFile(sslRoot + "cert_pwd")
	opts := append(DefaultServerTLSOptions(),
		WithVerifyPeer(true),
		WithVerifyHostName(false),
		WithVersion(ParseSSLProtocol("TLSv1.0"), tls.VersionTLS12),
		WithCipherSuits(ParseDefaultSSLCipherSuites("TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384")),
		WithKeyPass(string(pw)),
		WithCA(sslRoot+"trust.cer"),
		WithCert(sslRoot+"server.cer"),
		WithKey(sslRoot+"server_key.pem"),
	)
	clientTLSConfig, err := GetClientTLSConfig(opts...)
	if err != nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if len(clientTLSConfig.Certificates) == 0 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.RootCAs == nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if len(clientTLSConfig.CipherSuites) != 2 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.MinVersion != tls.VersionTLS10 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.MaxVersion != tls.VersionTLS12 {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.InsecureSkipVerify != true {
		t.Fatalf("GetClientTLSConfig failed")
	}

	// verify peer and peer host
	opts = append(opts,
		WithVerifyPeer(false),
		WithVerifyHostName(true),
	)
	clientTLSConfig, err = GetClientTLSConfig(opts...)
	if err != nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.RootCAs != nil || !clientTLSConfig.InsecureSkipVerify {
		t.Fatalf("GetClientTLSConfig failed")
	}
	opts = append(opts,
		WithVerifyPeer(true),
		WithVerifyHostName(false),
	)
	clientTLSConfig, err = GetClientTLSConfig(opts...)
	if err != nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.RootCAs == nil || !clientTLSConfig.InsecureSkipVerify {
		t.Fatalf("GetClientTLSConfig failed")
	}
	opts = append(opts,
		WithVerifyPeer(true),
		WithVerifyHostName(true),
	)
	clientTLSConfig, err = GetClientTLSConfig(opts...)
	if err != nil {
		t.Fatalf("GetClientTLSConfig failed")
	}
	if clientTLSConfig.RootCAs == nil || clientTLSConfig.InsecureSkipVerify {
		t.Fatalf("GetClientTLSConfig failed")
	}
}
