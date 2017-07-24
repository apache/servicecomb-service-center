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
package rest

import (
	"github.com/astaxie/beego"
	"github.com/servicecomb/service-center/pkg/common"
	"github.com/servicecomb/service-center/util"
	"net/http"
	"time"
)

var defaultRESTfulServer *http.Server

func ListenAndServeTLS(addr string, handler http.Handler) error {
	verifyClient := common.GetServerSSLConfig().VerifyClient
	tlsConfig, err := GetServerTLSConfig(verifyClient)
	if err != nil {
		return err
	}
	readTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("read_timeout", "5s"))
	writeTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("write_timeout", "5s"))
	maxHeaderBytes := beego.AppConfig.DefaultInt("max_header_bytes", 16384)
	defaultRESTfulServer = &http.Server{
		Addr:           addr,
		Handler:        handler,
		TLSConfig:      tlsConfig,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	util.LOGGER.Warnf(nil, "listen on server %s.", addr)
	// 证书已经在config里加载，这里不需要再重新加载
	return defaultRESTfulServer.ListenAndServeTLS("", "")
}
func ListenAndServe(addr string, handler http.Handler) error {
	readTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("read_timeout", "5s"))
	writeTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("write_timeout", "5s"))
	maxHeaderBytes := beego.AppConfig.DefaultInt("max_header_bytes", 16384)
	defaultRESTfulServer = &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	util.LOGGER.Warnf(nil, "listen on server %s.", addr)
	return defaultRESTfulServer.ListenAndServe()
}

func CloseServer() {
	if defaultRESTfulServer != nil {
		err := defaultRESTfulServer.Close()
		if err != nil {
			util.LOGGER.Errorf(err, "close RESTful server failed.")
		}
	}
}
