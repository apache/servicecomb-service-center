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
	"golang.org/x/net/context"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func GetIPFromContext(ctx context.Context) string {
	v, ok := FromContext(ctx, "x-remote-ip").(string)
	if !ok {
		return ""
	}
	return v
}

func UrlEncode(keys map[string]string) string {
	l := len(keys)
	if l == 0 {
		return ""
	}
	arr := make([]string, 0, l)
	for k, v := range keys {
		arr = append(arr, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	return StringJoin(arr, "&")
}

func ParseEndpoint(ep string) (string, error) {
	u, err := url.Parse(ep)
	if err != nil {
		return "", err
	}
	port := u.Port()
	if len(port) > 0 {
		return u.Hostname() + ":" + port, nil
	}
	return u.Hostname(), nil
}

func GetRealIP(r *http.Request) string {
	for _, h := range [2]string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		for _, ip := range addresses {
			ip = strings.TrimSpace(ip)
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() {
				continue
			}
			return ip
		}
	}
	addrs := strings.Split(r.RemoteAddr, ":")
	if len(addrs) > 0 {
		return addrs[0]
	}
	return ""
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func InetNtoIP(ipnr uint32) net.IP {
	return net.IPv4(byte(ipnr>>24), byte(ipnr>>16), byte(ipnr>>8), byte(ipnr))
}

func InetNtoa(ipnr uint32) string {
	return InetNtoIP(ipnr).String()
}

func InetAton(ip string) (ipnr uint32) {
	bytes := net.ParseIP(ip).To4()
	for i := 0; i < len(bytes); i++ {
		ipnr |= uint32(bytes[i])
		if i < 3 {
			ipnr <<= 8
		}
	}
	return
}

func ParseRequestURL(r *http.Request) string {
	if len(r.URL.Scheme) > 0 {
		return r.URL.String()
	}

	scheme := "https://"
	if r.TLS == nil {
		scheme = "http://"
	}
	return scheme + r.Host + r.RequestURI
}
