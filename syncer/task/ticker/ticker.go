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
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/task"
	"github.com/pkg/errors"
)

const (
	TaskName    = "ticker"
	intervalKey = "interval"
)

func init() {
	task.RegisterTasker(TaskName, NewTicker)
}

// Ticker struct
type Ticker struct {
	interval time.Duration
	handler  func()
	running  *utils.AtomicBool
	ticker   *time.Ticker
}

// NewTicker returns a ticker as a tasker
func NewTicker(params map[string]string) (task.Tasker, error) {
	val, ok := params[intervalKey]
	if !ok {
		return nil, errors.New("ticker: param interval notfound")
	}

	interval, err := time.ParseDuration(val)
	if err != nil {
		return nil, errors.Wrap(err, "ticker: parse interval duration failed")
	}

	return &Ticker{
		interval: interval,
		running:  utils.NewAtomicBool(false),
	}, nil
}

// Run ticker task
func (t *Ticker) Run(ctx context.Context) {
	t.running.DoToReverse(false, func() {
		t.ticker = time.NewTicker(t.interval)
		t.handler()
		go t.wait(ctx)
	})
}

// Handle ticker task
func (t *Ticker) Handle(handler func()) {
	t.handler = handler
}

func (t *Ticker) wait(ctx context.Context) {
	for {
		select {
		case <-t.ticker.C:
			t.handler()
		case <-ctx.Done():
			t.stop()
			return
		}
	}
}

// stop ticker task
func (t *Ticker) stop() {
	t.running.DoToReverse(true, func() {
		if t.ticker != nil {
			t.ticker.Stop()
		}
		log.Info("ticker stop")
	})

	log.Info("ticker done")
}
