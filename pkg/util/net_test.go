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
package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ip1 = "127.0.0.1"       // 2130706433
	ip2 = "0.0.0.0"         // 0
	ip3 = "255.255.255.255" // 4294967295
	n1  = 2130706433        // "127.0.0.1"
	n2  = 0                 // "0.0.0.0"
	n3  = 4294967295        // "255.255.255.255"
)

func TestInetAton(t *testing.T) {
	i := InetAton(ip1)
	if i != 2130706433 {
		t.Fatalf("InetAton(%s) error", ip1)
	}
	i = InetAton(ip2)
	if i != 0 {
		t.Fatalf("InetAton(%s) error", ip2)
	}
	i = InetAton(ip3)
	if i != 4294967295 {
		t.Fatalf("InetAton(%s) error", ip3)
	}
}

func TestInetNtoa(t *testing.T) {
	ip := InetNtoa(n1)
	if ip != ip1 {
		t.Fatalf("InetNtoa(%d) error", n1)
	}
	ip = InetNtoa(n2)
	if ip != ip2 {
		t.Fatalf("InetNtoa(%d) error", n2)
	}
	ip = InetNtoa(n3)
	if ip != ip3 {
		t.Fatalf("InetNtoa(%d) error", n3)
	}
}

func TestParseIpPort(t *testing.T) {
	ipPort := ParseIPPort("0.0.0.0")
	if ipPort.IP != "0.0.0.0" || ipPort.Port != 0 {
		t.Fatalf("ParseIPPort(0.0.0.0) error")
	}
	ipPort = ParseIPPort("0.0.0.0:1")
	if ipPort.IP != "0.0.0.0" || ipPort.Port != 1 {
		t.Fatalf("ParseIPPort(0.0.0.0) error")
	}
}

func TestParseEndpoint(t *testing.T) {
	ep, err := ParseEndpoint("")
	if err != nil || len(ep) > 0 {
		t.Fatalf("ParseEndpoint(\"\") failed, err = %s, ep = %s", err, ep)
	}
	ep, err = ParseEndpoint(":sssss")
	if err == nil || len(ep) > 0 {
		t.Fatalf("ParseEndpoint(\":sssss\") failed, err = %s, ep = %s", err, ep)
	}
	ep, err = ParseEndpoint("rest://127.0.0.1/?a=b")
	if err != nil || ep != "127.0.0.1" {
		t.Fatalf("ParseEndpoint(\"rest://127.0.0.1/?a=b\") failed, err = %s, ep = %s", err, ep)
	}
	ep, err = ParseEndpoint("rest://127.0.0.1:30100/?a=b")
	if err != nil || ep != "127.0.0.1:30100" {
		t.Fatalf("ParseEndpoint(\"rest://127.0.0.1:30100/?a=b\") failed, err = %s, ep = %s", err, ep)
	}
	ep, err = ParseEndpoint("rest://[2400:A480:AAAA:200::159]:30100/?a=b")
	if err != nil || ep != "[2400:A480:AAAA:200::159]:30100" {
		t.Fatalf("ParseEndpoint(\"rest://[2400:A480:AAAA:200::159]:30100/?a=b\") failed, err = %s, ep = %s", err, ep)
	}
	ep, err = ParseEndpoint("rest://[fe80::f816:3eff:fe17:c38b%25eht0]:30100/?a=b")
	if err != nil || ep != "[fe80::f816:3eff:fe17:c38b%eht0]:30100" {
		t.Fatalf("ParseEndpoint(\"rest://[fe80::f816:3eff:fe17:c38b%%25eht0]:30100/?a=b\") failed, err = %s, ep = %s", err, ep)
	}
}

func TestParseRequestURL(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:30100/x/?a=b&c=d#e", nil)
	url := ParseRequestURL(req)
	if url != "https://127.0.0.1:30100/x/?a=b&c=d#e" {
		t.Fatalf("TestParseRequestURL failed")
	}
	req.URL.Scheme = ""
	req.Host = "127.0.0.1:30100"
	req.RequestURI = "/x/?a=b&c=d#e"
	url = ParseRequestURL(req)
	if url != "http://127.0.0.1:30100/x/?a=b&c=d#e" {
		t.Fatalf("TestParseRequestURL failed")
	}
}

func TestGetRealIP(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://127.0.0.1:30100/x/?a=b&c=d#e", nil)
	req.RemoteAddr = "127.0.0.1:30100"
	ip := GetRealIP(req)
	assert.Equal(t, "127.0.0.1", ip)

	req.Header.Set("X-Real-Ip", "255.255.255.255")
	ip = GetRealIP(req)
	assert.Equal(t, "127.0.0.1", ip)

	req.Header.Set("X-Real-Ip", "4.4.4.4")
	ip = GetRealIP(req)
	assert.Equal(t, "4.4.4.4", ip)

	req.Header.Set("X-Forwarded-For", "1.1.1.1, 2.2.2.2, 3.3.3.3")
	ip = GetRealIP(req)
	assert.Equal(t, "1.1.1.1", ip)

	// ipv6
	req, _ = http.NewRequest(http.MethodGet, "https://127.0.0.1:30100/x/?a=b&c=d#e", nil)
	req.RemoteAddr = "[::1]:30100"
	ip = GetRealIP(req)
	assert.Equal(t, "::1", ip)

	req.RemoteAddr = "[2008:0:0:0:8:800:200C:417A]:30100"
	ip = GetRealIP(req)
	assert.Equal(t, "2008:0:0:0:8:800:200C:417A", ip)
}
