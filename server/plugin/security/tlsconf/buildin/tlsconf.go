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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/security/cipher"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	"github.com/go-chassis/foundation/tlsutil"
)

var (
	clientTLSConfig *tls.Config
	serverTLSConfig *tls.Config
	mux             sync.Mutex
)

func GetSSLPath(path string) string {
	dir := tlsconf.GetOptions().Dir
	if len(dir) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, "etc", "ssl", path)
	}
	return os.ExpandEnv(filepath.Join(dir, path))
}

func GetPassphase() (decrypt string) {
	passphase, err := os.ReadFile(GetSSLPath("cert_pwd"))
	if err != nil {
		log.Error("read file cert_pwd failed.", err)
	}

	decrypt = util.BytesToStringWithNoCopy(passphase)
	if len(decrypt) > 0 {
		tmp, err := cipher.Decrypt(decrypt)
		if err != nil {
			log.Error(fmt.Sprintf("decrypt ssl passphase(%d) failed.", len(decrypt)), err)
		} else {
			decrypt = tmp
		}
	}
	return decrypt
}

func GetClientTLSConfig() (_ *tls.Config, err error) {
	mux.Lock()
	defer mux.Unlock()
	if clientTLSConfig != nil {
		return clientTLSConfig, nil
	}

	passphase := GetPassphase()

	opts := append(tlsutil.DefaultClientTLSOptions(),
		tlsutil.WithVerifyPeer(tlsconf.GetOptions().VerifyPeer),
		tlsutil.WithVerifyHostName(false),
		tlsutil.WithVersion(tlsutil.ParseSSLProtocol(tlsconf.GetOptions().ClientMinVersion), tlsutil.MaxSupportedTLSVersion),
		tlsutil.WithCipherSuits(tlsutil.ParseDefaultSSLCipherSuites(tlsconf.GetOptions().ClientCiphers)),
		tlsutil.WithKeyPass(passphase),
		tlsutil.WithCA(GetSSLPath("trust.cer")),
		tlsutil.WithCert(GetSSLPath("server.cer")),
		tlsutil.WithKey(GetSSLPath("server_key.pem")),
	)
	clientTLSConfig, err = tlsutil.GetClientTLSConfig(opts...)

	if clientTLSConfig != nil {
		log.Info(fmt.Sprintf("client ssl configs enabled, verifyclient %t, minv %#x, cipers %d, pphase %d.",
			tlsconf.GetOptions().VerifyPeer,
			clientTLSConfig.MinVersion,
			len(clientTLSConfig.CipherSuites),
			len(passphase)))
	}
	return clientTLSConfig, err
}

func GetServerTLSConfig() (_ *tls.Config, err error) {
	mux.Lock()
	defer mux.Unlock()
	if serverTLSConfig != nil {
		return serverTLSConfig, nil
	}

	passphase := GetPassphase()

	opts := append(tlsutil.DefaultServerTLSOptions(),
		tlsutil.WithVerifyPeer(tlsconf.GetOptions().VerifyPeer),
		tlsutil.WithVersion(tlsutil.ParseSSLProtocol(tlsconf.GetOptions().MinVersion), tlsutil.MaxSupportedTLSVersion),
		tlsutil.WithCipherSuits(tlsutil.ParseDefaultSSLCipherSuites(tlsconf.GetOptions().Ciphers)),
		tlsutil.WithKeyPass(passphase),
		tlsutil.WithCA(GetSSLPath("trust.cer")),
		tlsutil.WithCert(GetSSLPath("server.cer")),
		tlsutil.WithKey(GetSSLPath("server_key.pem")),
	)

	serverTLSConfig, err = tlsutil.GetServerTLSConfig(opts...)

	if serverTLSConfig != nil {
		log.Info(fmt.Sprintf("server ssl configs enabled, verifyClient %t, minv %#x, ciphers %d, phase %d.",
			tlsconf.GetOptions().VerifyPeer,
			serverTLSConfig.MinVersion,
			len(serverTLSConfig.CipherSuites),
			len(passphase)))
	}
	return serverTLSConfig, err
}
