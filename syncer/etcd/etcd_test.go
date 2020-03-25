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

package etcd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestETCDServer(t *testing.T) {
	defer os.RemoveAll("test-data")
	svr, err := NewServer(
		WithName("test"),
		WithDataDir("test-data/a"),
		WithPeerAddr("127.0.0.1:8090"),
	)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	err = startServer(ctx, svr)
	assert.Nil(t, err)

	svr.IsLeader()
	svr.Storage()

	err = svr.AddOptions(WithAddPeers("defalt", "127.0.0.1:8092"))
	cancel()
	assert.NotNil(t, err)

	svr.Stop()

	svr, err = NewServer(
		WithName("test"),
		WithDataDir("test-data/b"),
		WithPeerAddr("127.0.0.1:8091"),
	)
	assert.Nil(t, err)

	err = svr.AddOptions(WithAddPeers("defalt", "127.0.0.1:8092"))
	assert.Nil(t, err)

	ctx, cancel = context.WithCancel(context.Background())
	err = startServer(ctx, svr)
	cancel()
	assert.NotNil(t, err)

	<-time.After(time.Second * 3)
}

func startServer(ctx context.Context, svr *Server) (err error) {
	svr.Start(ctx)
	select {
	case <-svr.Ready():
	case <-svr.Stopped():
		err = errors.New("start etcd server failed")
	case <-time.After(time.Second * 3):
		err = errors.New("start etcd server timeout")
	}
	return
}
