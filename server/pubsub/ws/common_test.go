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
	"errors"
	"testing"

	"github.com/apache/servicecomb-service-center/server/pubsub/ws"
	"github.com/stretchr/testify/assert"
)

func TestSendEstablishError(t *testing.T) {
	mock := NewTest()
	t.Run("should read the err when call", func(t *testing.T) {
		ws.SendEstablishError(mock.ServerConn, errors.New("error"))
		_, message, err := mock.ClientConn.ReadMessage()
		assert.Nil(t, err)
		assert.Equal(t, "error", string(message))
	})
}

func TestWatch(t *testing.T) {
	t.Run("should return when ctx cancelled", func(t *testing.T) {
		mock := NewTest()
		mock.ServerConn.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ws.Watch(ctx, "", mock.ServerConn)
	})
}
