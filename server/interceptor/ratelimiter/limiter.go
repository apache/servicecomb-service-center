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
	"errors"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/pkg/httplimiter"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Limiter struct {
	conns int64

	httpLimiter *httplimiter.HttpLimiter
}

var limiter *Limiter
var mux sync.Mutex

func GetLimiter() *Limiter {
	if limiter == nil {
		mux.Lock()
		if limiter == nil {
			limiter = new(Limiter)
			limiter.LoadConfig()
		}
		mux.Unlock()
	}

	return limiter
}

func (this *Limiter) LoadConfig() {
	ttl := time.Second
	switch core.ServerInfo.Config.LimitTTLUnit {
	case "ms":
		ttl = time.Millisecond
	case "m":
		ttl = time.Minute
	case "h":
		ttl = time.Hour
	}
	this.conns = core.ServerInfo.Config.LimitConnections
	this.httpLimiter = httplimiter.NewHttpLimiter(this.conns, ttl)
	iplookups := core.ServerInfo.Config.LimitIPLookup
	this.httpLimiter.IPLookups = strings.Split(iplookups, ",")

	util.Logger().Warnf(nil, "Rate-limit Load config, ttl: %s, conns: %d, iplookups: %s", ttl, this.conns, iplookups)
}

func (this *Limiter) Handle(w http.ResponseWriter, r *http.Request) error {
	if this.conns <= 0 {
		return nil
	}

	httplimiter.SetResponseHeaders(this.httpLimiter, w)
	httpError := httplimiter.LimitByRequest(this.httpLimiter, r)
	if httpError != nil {
		w.Header().Add("Content-Type", this.httpLimiter.ContentType)
		w.WriteHeader(httpError.StatusCode)
		w.Write(util.StringToBytesWithNoCopy(httpError.Message))
		util.Logger().Warn("Reached maximum request limit!", nil)
		return errors.New(httpError.Message)
	}
	return nil
}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	return GetLimiter().Handle(w, r)
}
