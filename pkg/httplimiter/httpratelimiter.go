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

package httplimiter

import (
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/ratelimiter"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HTTPErrorMessage struct {
	Message    string
	StatusCode int
}

func (httpErrorMessage *HTTPErrorMessage) Error() string {
	return fmt.Sprintf("%v: %v", httpErrorMessage.StatusCode, httpErrorMessage.Message)
}

type HttpLimiter struct {
	HttpMessage    string
	ContentType    string
	StatusCode     int
	RequestLimit   int64
	TTL            time.Duration
	IPLookups      []string
	Methods        []string
	Headers        map[string][]string
	BasicAuthUsers []string
	leakyBuckets   map[string]*ratelimiter.LeakyBucket
	sync.RWMutex
}

func LimitBySegments(limiter *HttpLimiter, keys []string) *HTTPErrorMessage {
	if limiter.LimitExceeded(strings.Join(keys, "|")) {
		return &HTTPErrorMessage{Message: limiter.HttpMessage, StatusCode: limiter.StatusCode}
	}

	return nil
}

func LimitByRequest(httpLimiter *HttpLimiter, r *http.Request) *HTTPErrorMessage {
	sliceKeys := BuildSegments(httpLimiter, r)

	for _, keys := range sliceKeys {
		httpError := LimitBySegments(httpLimiter, keys)
		if httpError != nil {
			return httpError
		}
	}

	return nil
}

func BuildSegments(httpLimiter *HttpLimiter, r *http.Request) [][]string {
	remoteIP := getRemoteIP(httpLimiter.IPLookups, r)
	urlPath := r.URL.Path
	sliceKeys := make([][]string, 0)

	if remoteIP == "" {
		return sliceKeys
	}

	if httpLimiter.Methods != nil && httpLimiter.Headers != nil && httpLimiter.BasicAuthUsers != nil {
		if checkExistence(httpLimiter.Methods, r.Method) {
			for headerKey, headerValues := range httpLimiter.Headers {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					username, _, ok := r.BasicAuth()
					if ok && checkExistence(httpLimiter.BasicAuthUsers, username) {
						sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method, headerKey, username})
					}

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					for _, headerValue := range headerValues {
						username, _, ok := r.BasicAuth()
						if ok && checkExistence(httpLimiter.BasicAuthUsers, username) {
							sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method, headerKey, headerValue, username})
						}
					}
				}
			}
		}

	} else if httpLimiter.Methods != nil && httpLimiter.Headers != nil {
		if checkExistence(httpLimiter.Methods, r.Method) {
			for headerKey, headerValues := range httpLimiter.Headers {
				if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
					sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method, headerKey})

				} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
					for _, headerValue := range headerValues {
						sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method, headerKey, headerValue})
					}
				}
			}
		}

	} else if httpLimiter.Methods != nil && httpLimiter.BasicAuthUsers != nil {
		if checkExistence(httpLimiter.Methods, r.Method) {
			username, _, ok := r.BasicAuth()
			if ok && checkExistence(httpLimiter.BasicAuthUsers, username) {
				sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method, username})
			}
		}

	} else if httpLimiter.Methods != nil {
		if checkExistence(httpLimiter.Methods, r.Method) {
			sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, r.Method})
		}

	} else if httpLimiter.Headers != nil {
		for headerKey, headerValues := range httpLimiter.Headers {
			if (headerValues == nil || len(headerValues) <= 0) && r.Header.Get(headerKey) != "" {
				sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, headerKey})

			} else if len(headerValues) > 0 && r.Header.Get(headerKey) != "" {
				for _, headerValue := range headerValues {
					sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, headerKey, headerValue})
				}
			}
		}

	} else if httpLimiter.BasicAuthUsers != nil {
		username, _, ok := r.BasicAuth()
		if ok && checkExistence(httpLimiter.BasicAuthUsers, username) {
			sliceKeys = append(sliceKeys, []string{remoteIP, urlPath, username})
		}
	} else {
		sliceKeys = append(sliceKeys, []string{remoteIP, urlPath})
	}

	return sliceKeys
}

func SetResponseHeaders(limiter *HttpLimiter, w http.ResponseWriter) {
	w.Header().Add("X-Rate-Limit-Limit", strconv.FormatInt(limiter.RequestLimit, 10))
	w.Header().Add("X-Rate-Limit-Duration", limiter.TTL.String())
}

func checkExistence(sliceString []string, needle string) bool {
	for _, b := range sliceString {
		if b == needle {
			return true
		}
	}
	return false
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getRemoteIP(ipLookups []string, r *http.Request) string {
	realIP := r.Header.Get("X-Real-IP")
	forwardedFor := r.Header.Get("X-Forwarded-For")

	for _, lookup := range ipLookups {
		if lookup == "RemoteAddr" {
			return ipAddrFromRemoteAddr(r.RemoteAddr)
		}
		if lookup == "X-Forwarded-For" && forwardedFor != "" {
			parts := strings.Split(forwardedFor, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}
			return parts[0]
		}
		if lookup == "X-Real-IP" && realIP != "" {
			return realIP
		}
	}

	return ""
}

func NewHttpLimiter(max int64, ttl time.Duration) *HttpLimiter {
	limiter := &HttpLimiter{RequestLimit: max, TTL: ttl}
	limiter.ContentType = "text/plain; charset=utf-8"
	limiter.HttpMessage = "You have reached maximum request limit."
	limiter.StatusCode = http.StatusTooManyRequests
	limiter.leakyBuckets = make(map[string]*ratelimiter.LeakyBucket)
	limiter.IPLookups = []string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}

	return limiter
}

func (rateLimiter *HttpLimiter) LimitExceeded(key string) bool {
	rateLimiter.Lock()
	if _, found := rateLimiter.leakyBuckets[key]; !found {
		rateLimiter.leakyBuckets[key] = ratelimiter.NewLeakyBucket(rateLimiter.TTL, rateLimiter.RequestLimit, rateLimiter.RequestLimit)
	}
	_, isInLimits := rateLimiter.leakyBuckets[key].MaximumTakeDuration(1, 0)
	rateLimiter.Unlock()
	if isInLimits {
		return false
	}
	return true
}
