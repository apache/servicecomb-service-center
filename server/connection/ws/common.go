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

package ws

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/connection"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/gorilla/websocket"
)

func ListAndWatch(ctx context.Context, serviceID string, f func() ([]*registry.WatchInstanceResponse, int64), conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)
	socket := New(ctx, conn, event.NewInstanceSubscriber(serviceID, domainProject, f))

	connection.ReportSubscriber(domain, Websocket, 1)
	process(socket)
	connection.ReportSubscriber(domain, Websocket, -1)
}

func process(socket *WebSocket) {
	if err := socket.Init(); err != nil {
		return
	}

	socket.HandleControlMessage()

	socket.Stop()
}

func SendEstablishError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	log.Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		log.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}
