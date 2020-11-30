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
package ticker

import (
	"context"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

// TaskTicker task of ticker struct
type TaskTicker struct {
	interval time.Duration
	handler  func(ctx context.Context)
	once     sync.Once
	ticker   *time.Ticker
}

// NewTaskTicker new task ticker with interval
func NewTaskTicker(interval int, handler func(ctx context.Context)) *TaskTicker {
	return &TaskTicker{
		interval: time.Second * time.Duration(interval),
		handler:  handler,
	}
}

// Start start task ticker
func (t *TaskTicker) Start(ctx context.Context) {
	t.once.Do(func() {
		t.handler(ctx)
		t.ticker = time.NewTicker(t.interval)
		for {
			select {
			case <-t.ticker.C:
				t.handler(ctx)
			case <-ctx.Done():
				t.Stop()
				return
			}
		}
	})
}

// Start stop task ticker
func (t *TaskTicker) Stop() {
	if t.ticker == nil {
		return
	}
	t.ticker.Stop()
	log.Info("ticker stop")
	t.ticker = nil
}
