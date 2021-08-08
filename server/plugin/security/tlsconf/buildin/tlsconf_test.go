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
package buildin_test

import (
	"crypto/tls"
	"testing"

	_ "github.com/apache/servicecomb-service-center/server/plugin/security/cipher/buildin"

	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf/buildin"
	"github.com/stretchr/testify/assert"
)

func init() {
	config.Init()
	_ = tlsconf.Init(tlsconf.Options{
		Dir:              "../../../../../examples/service_center/ssl/",
		MinVersion:       "TLSv1.2",
		ClientMinVersion: "TLSv1.2",
		VerifyPeer:       true,
	})
}

func TestGetServerTLSConfig(t *testing.T) {
	serverTLSConfig, err := buildin.GetServerTLSConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, serverTLSConfig.Certificates)
	assert.NotNil(t, serverTLSConfig.ClientCAs)
	assert.Empty(t, serverTLSConfig.CipherSuites)
	assert.True(t, tls.VersionTLS12 == serverTLSConfig.MinVersion)
	assert.Equal(t, int(serverTLSConfig.MaxVersion), tls.VersionTLS13)
	assert.Equal(t, serverTLSConfig.ClientAuth, tls.RequireAndVerifyClientCert)
}

func TestGetClientTLSConfig(t *testing.T) {
	clientTLSConfig, err := buildin.GetClientTLSConfig()
	assert.NoError(t, err)
	assert.NotEmpty(t, clientTLSConfig.Certificates)
	assert.NotNil(t, clientTLSConfig.RootCAs)
	assert.Empty(t, clientTLSConfig.CipherSuites)
	assert.True(t, tls.VersionTLS12 == clientTLSConfig.MinVersion)
	assert.Equal(t, int(clientTLSConfig.MaxVersion), tls.VersionTLS13)
	assert.True(t, clientTLSConfig.InsecureSkipVerify)
}
