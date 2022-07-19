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

package istioconnector

import (
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestDebouncing(t *testing.T) {
	pushed := 0
	eventQueue := 0
	debouncedFn := debounce(func() {
		pushed++
		eventQueue = 0
	}, utils.PUSH_DEBOUNCE_INTERVAL, utils.PUSH_DEBOUNCE_MAX_INTERVAL)

	debouncer := func() {
		eventQueue++
		debouncedFn()
	}

	t.Run("add an event to the event queue, the event will be hold for 1 sec in the queue, if no more event comes in within 1 sec, then push it", func(t *testing.T) {
		debouncer()
		assert.Equal(t, 0, pushed)
		assert.Equal(t, 1, eventQueue)

		time.Sleep(time.Second * 2)
		assert.Equal(t, 1, pushed)
	})

	t.Run("add an event to the event queue, update the event after 500ms, the event should only be pushed once after 1 sec", func(t *testing.T) {
		pushed = 0
		debouncer()
		time.Sleep(time.Millisecond * 500)
		debouncer()
		time.Sleep(time.Second * 2)
		assert.Equal(t, 1, pushed)
	})

	t.Run("continuously inserting events, each event 500s apart, only the latest event will be pushed", func(t *testing.T) {
		pushed = 0
		for i := 0; i < 20; i++ {
			debouncer()
			time.Sleep(time.Millisecond * 500)
		}
		assert.Equal(t, 1, pushed)
	})

	t.Run("continuously inserting events, each event 1500s apart, will trigger push operation for 20 times", func(t *testing.T) {
		pushed = 0
		for i := 0; i < 20; i++ {
			debouncer()
			time.Sleep(time.Millisecond * 1500)
		}
		assert.Equal(t, 20, pushed)
	})

}
