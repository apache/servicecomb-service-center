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

package syncer

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strings"
)

var (
	ServiceAPI   = &Service{}
	configs      map[string]string
	environments = make(map[string]string)
)

func init() {
	// cache envs
	for _, kv := range os.Environ() {
		arr := strings.Split(kv, "=")
		environments[arr[0]] = arr[1]
	}

	// cache configs
	configs, _ = beego.AppConfig.GetSection("default")
	if section, err := beego.AppConfig.GetSection(beego.BConfig.RunMode); err == nil {
		for k, v := range section {
			configs[k] = v
		}
	}
}

type Service struct {
}

func (service *Service) WatchInstance(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade failed", err)
		return
	}
	defer conn.Close()

	syncernotify.DoWebSocketWatch(context.Background(), conn)

}
