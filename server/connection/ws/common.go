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
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/gorilla/websocket"
)

func Watch(ctx context.Context, serviceID string, conn *websocket.Conn) {
	domainProject := util.ParseDomainProject(ctx)
	domain := util.ParseDomain(ctx)

	ws := NewWebSocket(domainProject, serviceID, conn)
	HealthChecker().Accept(ws)

	subscriber := event.NewInstanceSubscriber(serviceID, domainProject)
	err := event.Center().AddSubscriber(subscriber)
	if err != nil {
		SendEstablishError(conn, err)
		return
	}

	metrics.ReportSubscriber(domain, Websocket, 1)
	defer metrics.ReportSubscriber(domain, Websocket, -1)

	pool := gopool.New(ctx).Do(func(ctx context.Context) {
		if err := NewBroker(ws, subscriber).Listen(ctx); err != nil {
			log.Error(fmt.Sprintf("[%s] listen service[%s] failed", conn.RemoteAddr(), serviceID), err)
		}
	})
	defer pool.Done()

	if err := ws.ReadMessage(); err != nil {
		log.Error(fmt.Sprintf("read subscriber[%s][%s] message failed", serviceID, conn.RemoteAddr()), err)
		subscriber.SetError(err)
	}
}

func SendEstablishError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	log.Errorf(err, "establish[%s] websocket watch failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		log.Errorf(err, "establish[%s] websocket watch failed: write message failed.", remoteAddr)
	}
}
