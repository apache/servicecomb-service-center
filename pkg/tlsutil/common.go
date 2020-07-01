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
)

var TLSCipherSuiteMap = map[string]uint16{
	"TLS_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":       tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
}

var TLSVersionMap = map[string]uint16{
	"TLSv1.0": tls.VersionTLS10,
	"TLSv1.1": tls.VersionTLS11,
	"TLSv1.2": tls.VersionTLS12,
}

var cipherSuite []uint16

// MaxSupportedTLSVersion is the max supported TLS version
var MaxSupportedTLSVersion uint16 = tls.VersionTLS12

func TLSCipherSuits() []uint16 {
	if cipherSuite != nil {
		return cipherSuite
	}
	for _, c := range TLSCipherSuiteMap {
		cipherSuite = append(cipherSuite, c)
	}
	return cipherSuite
}
