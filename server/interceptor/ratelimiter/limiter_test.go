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
package ratelimiter

import (
	"github.com/didip/tollbooth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"time"
)

var _ = Describe("HttpLimiter", func() {
	var (
		limiter *Limiter
	)

	BeforeEach(func() {
		limiter = new(Limiter)
		limiter.LoadConfig()
	})
	Describe("LoadConfig", func() {
		Context("Normal", func() {
			It("should be ok", func() {
				Expect(limiter.conns).To(Equal(int64(0)))
				res := []string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}
				for i, val := range limiter.httpLimiter.IPLookups {
					Expect(val).To(Equal(res[i]))
				}
			})
		})
	})
	Describe("FuncHandler", func() {
		var ts *httptest.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := limiter.Handle(w, r)
			if err != nil {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Testing..."))
		}))
		Context("Connections > 0", func() {
			It("should not be router", func() {
				limiter.conns = 1
				limiter.httpLimiter = tollbooth.NewLimiter(1, time.Second)
				resp, err := http.Get(ts.URL)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				resp, err = http.Get(ts.URL)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).ToNot(Equal(http.StatusOK))
			})
		})
		Context("Connections <= 0", func() {
			It("should be router", func() {
				limiter.conns = 0
				limiter.httpLimiter = tollbooth.NewLimiter(0, time.Second)
				resp, err := http.Get(ts.URL)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				resp, err = http.Get(ts.URL)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
