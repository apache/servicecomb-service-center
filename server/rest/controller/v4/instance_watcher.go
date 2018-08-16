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
package v4

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/gorilla/websocket"
	"net/http"
)

type WatchService struct {
	//
}

func (this *WatchService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/watcher", this.Watch},
		{rest.HTTP_METHOD_GET, "/v4/:project/registry/microservices/:serviceId/listwatcher", this.ListAndWatch},
	}
}

func upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		/*Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {

		  },*/
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade failed.", err)
		// controller.WriteError(w, scerr.ErrInternal, "Upgrade error")
	}
	return conn, err
}

func (this *WatchService) Watch(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	r.Method = "WATCH"
	core.InstanceAPI.WebSocketWatch(r.Context(), &pb.WatchInstanceRequest{
		SelfServiceId: r.URL.Query().Get(":serviceId"),
	}, conn)
}

func (this *WatchService) ListAndWatch(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	r.Method = "WATCHLIST"
	core.InstanceAPI.WebSocketListAndWatch(r.Context(), &pb.WatchInstanceRequest{
		SelfServiceId: r.URL.Query().Get(":serviceId"),
	}, conn)
}
