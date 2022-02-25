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

package client_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestNewSetForConfig(t *testing.T) {
	conn, err := rpc.GetPickFirstLbConn(
		&rpc.Config{
			Addrs:       []string{"127.0.0.1:30105"},
			Scheme:      "test",
			ServiceName: "serviceName",
		})
	assert.NoError(t, err)
	defer conn.Close()
	set := client.NewSet(conn)
	_, err = set.EventServiceClient.Sync(context.TODO(), &v1sync.EventList{Events: []*v1sync.Event{
		{Action: "create"},
	}})
	assert.NoError(t, err)
}
