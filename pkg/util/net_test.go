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

import "testing"

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
	ipPort := ParseIpPort("0.0.0.0")
	if ipPort.IP != "0.0.0.0" || ipPort.Port != 0 {
		t.Fatalf("ParseIpPort(0.0.0.0) error", n3)
	}
	ipPort = ParseIpPort("0.0.0.0:1")
	if ipPort.IP != "0.0.0.0" || ipPort.Port != 1 {
		t.Fatalf("ParseIpPort(0.0.0.0) error", n3)
	}
}
