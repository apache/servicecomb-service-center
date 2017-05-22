//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package ratelimiter

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/config"
	"github.com/servicecomb/service-center/util"
	"net/http"
	"strings"
	"time"
)

type Limiter struct {
	conns int64

	tbLimiter *config.Limiter
}

var limiter *Limiter

func GetLimiter() *Limiter {
	if limiter == nil {
		limiter = new(Limiter)
		limiter.LoadConfig()
	}
	return limiter
}

func (this *Limiter) LoadConfig() {
	ttl := time.Second
	switch beego.AppConfig.DefaultString("limit_ttl", "s") {
	case "ms":
		ttl = time.Millisecond
	case "m":
		ttl = time.Minute
	case "h":
		ttl = time.Hour
	}
	this.conns = int64(beego.AppConfig.DefaultInt("limit_conns", 0))
	this.tbLimiter = tollbooth.NewLimiter(this.conns, ttl)
	iplookups := beego.AppConfig.DefaultString("limit_iplookups",
		"RemoteAddr,X-Forwarded-For,X-Real-IP")
	this.tbLimiter.IPLookups = strings.Split(iplookups, ",")

	util.LOGGER.Warnf(nil, "Rate-limit Load config, ttl: %s, conns: %d, iplookups: %s", ttl, this.conns, iplookups)
}

func (this *Limiter) Handle(w http.ResponseWriter, r *http.Request) error {
	if this.conns <= 0 {
		return nil
	}

	tollbooth.SetResponseHeaders(this.tbLimiter, w)
	httpError := tollbooth.LimitByRequest(this.tbLimiter, r)
	if httpError != nil {
		w.Header().Add("Content-Type", this.tbLimiter.MessageContentType)
		w.WriteHeader(httpError.StatusCode)
		w.Write([]byte(httpError.Message))
		util.LOGGER.Warn("Reached maximum request limit!", nil)
		return errors.New(httpError.Message)
	}
	return nil
}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	return GetLimiter().Handle(w, r)
}
