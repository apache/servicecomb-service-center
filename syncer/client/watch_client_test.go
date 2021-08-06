/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package client

import (
	"fmt"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestWatchInstance(t *testing.T) {

	addressWithPro := "http://127.0.0.1:8989"
	addressWithoutPro := "127.0.0.1:8888"
	cli := NewWatchClient(addressWithPro)
	assert.Equal(t, addressWithPro, cli.addr)

	cli = NewWatchClient(addressWithoutPro)
	assert.Equal(t, addressWithoutPro, cli.addr)

	err := cli.WatchInstances(fakeAddToQueue)
	assert.Error(t, err)

	cli.WatchInstanceHeartbeat(fakeAddToQueue)

}

func fakeAddToQueue(event *dump.WatchInstanceChangedEvent) {
	log.Debug(fmt.Sprintf("success add instance event to queue: %v", event))
}
