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

package utils

import (
	"net"
	"strconv"
	"strings"
)

// SplitHostPort returns the parts of the address and port. If the port does not exist, use defaultPort.
func SplitHostPort(address string, defaultPort int) (string, int, error) {
	_, _, err := net.SplitHostPort(address)
	if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
		index := strings.LastIndexByte(address, ':')
		return SplitAddress(address[:index+1] + strconv.Itoa(defaultPort))
	}
	return ResolveAddr(address)
}

// SplitAddress returns the parts of the address and port.
func SplitAddress(address string) (string, int, error) {
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}
	return ResolveAddr(address)
}

// ResolveAddr Resolve the address with tcp.
func ResolveAddr(address string) (string, int, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return "", 0, err
	}

	return addr.IP.String(), addr.Port, nil
}
