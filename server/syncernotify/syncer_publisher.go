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

package syncernotify

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/alarm"
)

var publisher *Publisher

func init() {
	publisher = NewPublisher()
	publisher.Run()
}

type Publisher struct {
	ws        *WebSocket
	goroutine *gopool.Pool
}

func (wh *Publisher) Run() {
	gopool.Go(publisher.loop)
}

func (wh *Publisher) Stop() {
	wh.goroutine.Close(true)
}

func (wh *Publisher) dispatch(ws *WebSocket, payload interface{}) {
	wh.goroutine.Do(func(ctx context.Context) {
		ws.HandleWatchWebSocketJob(payload)
	})
}

func (wh *Publisher) loop(ctx context.Context) {
	defer wh.Stop()
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			// server shutdown
			return
		case <-ticker.C:
			wsIsActive := true
			if wh.ws != nil {
				if payload := wh.ws.Pick(); payload != nil {
					if _, ok := payload.(error); ok {
						wsIsActive = false
					}
					wh.dispatch(wh.ws, payload)
				}
			}
			if !wsIsActive {
				log.Debug(fmt.Sprintf("release websocket conn :%s", wh.ws.conn.RemoteAddr()))

				err := wh.ws.conn.Close()
				if err != nil {
					log.Error("conn close failed", err)
				}

				wh.ws = nil

				err = alarm.Raise(alarm.IDWebsocketOfScSyncerLost, alarm.AdditionalContext("%v", err))
				if err != nil {
					log.Error("alarm error", err)
				}
			}
		}
	}
}

func (wh *Publisher) Accept(ws *WebSocket) {
	log.Debug(fmt.Sprintf("get a new websocket:%s", ws.conn.RemoteAddr()))
	wh.ws = ws
}

func NewPublisher() *Publisher {
	return &Publisher{
		goroutine: gopool.New(context.Background()),
	}
}

func Instance() *Publisher {
	return publisher
}
