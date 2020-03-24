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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	params := map[string]string{}
	_, err := NewTicker(params)
	assert.NotNil(t, err)

	params[intervalKey] = "1ams"
	_, err = NewTicker(params)
	assert.NotNil(t, err)

	params[intervalKey] = "1s"

	task, err := NewTicker(params)
	assert.Nil(t, err)

	isRunning := false
	stopped := make(chan struct{})
	task.Handle(func() {
		if isRunning {
			close(stopped)
		} else {
			isRunning = true
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	go task.Run(ctx)

	<-stopped
	assert.True(t, isRunning)
	ctx.Done()
	cancel()
	time.Sleep(time.Second)
}
