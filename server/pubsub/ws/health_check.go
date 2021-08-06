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
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/foundation/gopool"
)

var checker *HealthCheck

func init() {
	checker = NewHealthCheck()
	checker.Run()
}

type HealthCheck struct {
	wss  []*WebSocket
	lock sync.Mutex
}

func (wh *HealthCheck) Run() {
	gopool.Go(checker.loop)
}

func (wh *HealthCheck) loop(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			// server shutdown
			return
		case <-ticker.C:
			for _, ws := range wh.wss {
				if t := ws.NeedCheck(); t == nil {
					continue
				}
				wh.check(ws)
			}
		}
	}
}

func (wh *HealthCheck) check(ws *WebSocket) {
	gopool.Go(func(ctx context.Context) {
		if err := ws.CheckHealth(ctx); err != nil {
			wh.Remove(ws)
			log.Error(fmt.Sprintf("checker removed unhealth websocket[%s]", ws.RemoteAddr), err)
		}
	})
}

func (wh *HealthCheck) Accept(ws *WebSocket) int {
	wh.lock.Lock()
	wh.wss = append(wh.wss, ws)
	n := len(wh.wss)
	wh.lock.Unlock()
	return n
}

func (wh *HealthCheck) Remove(ws *WebSocket) int {
	wh.lock.Lock()
	for i, t := range wh.wss {
		if t == ws {
			wh.wss = append(wh.wss[0:i], wh.wss[i+1:]...)
			break
		}
	}
	n := len(wh.wss)
	wh.lock.Unlock()
	return n
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func HealthChecker() *HealthCheck {
	return checker
}
