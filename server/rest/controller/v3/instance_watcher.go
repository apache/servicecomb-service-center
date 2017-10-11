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
package v3

import (
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/rest/controller"
	"github.com/gorilla/websocket"
	"net/http"
)

type WatchService struct {
	//
}

func (this *WatchService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_GET, "/registry/v3/microservices/:serviceId/watcher", this.Watch},
		{rest.HTTP_METHOD_GET, "/registry/v3/microservices/:serviceId/listwatcher", this.ListAndWatch},
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
		util.Logger().Error("upgrade failed.", err)
		controller.WriteText(http.StatusInternalServerError, "Upgrade error", w)
	}
	return conn, err
}

func (this *WatchService) Watch(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade(w, r)
	if err != nil {
		return
	}
	defer conn.Close()
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
	core.InstanceAPI.WebSocketListAndWatch(r.Context(), &pb.WatchInstanceRequest{
		SelfServiceId: r.URL.Query().Get(":serviceId"),
	}, conn)
}
