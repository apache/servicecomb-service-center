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

package ws_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/server/connection/ws"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/stretchr/testify/assert"
)

func TestNewBroker(t *testing.T) {
	t.Run("should not return nil when new broker", func(t *testing.T) {
		assert.NotNil(t, ws.NewBroker(nil, nil))

	})
}

func TestBroker_Listen(t *testing.T) {
	t.Run("should return err when listen context cancelled", func(t *testing.T) {
		broker := ws.NewBroker(nil, event.NewInstanceSubscriber("", ""))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.Equal(t, context.Canceled, broker.Listen(ctx))
	})
}
