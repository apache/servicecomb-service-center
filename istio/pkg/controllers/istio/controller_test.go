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

package istio

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

	debouncer()

	assert.Equal(t, 0, pushed)
	assert.Equal(t, 1, eventQueue)

	time.Sleep(time.Second * 2)
	assert.Equal(t, 1, pushed)

	pushed = 0
	debouncer()
	time.Sleep(time.Millisecond * 500)
	debouncer()
	time.Sleep(time.Second * 2)
	assert.Equal(t, 1, pushed)

	pushed = 0
	for i := 0; i < 20; i++ {
		debouncer()
		time.Sleep(time.Millisecond * 500)
	}
	assert.Equal(t, 1, pushed)

	pushed = 0
	for i := 0; i < 20; i++ {
		debouncer()
		time.Sleep(time.Millisecond * 1500)
	}
	assert.Equal(t, 20, pushed)
}
