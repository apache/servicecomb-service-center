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

package serf

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
	"github.com/stretchr/testify/assert"
)

func TestSerfServer(t *testing.T) {
	svr := defaultServer()

	ctx, cancel := context.WithCancel(context.Background())
	err := startServer(ctx, svr)
	assert.Nil(t, err)

	err = svr.UserEvent("test-event", []byte("test-data"))
	assert.Nil(t, err)

	_, err = svr.Join([]string{"127.0.0.1:35151"})
	assert.Nil(t, err)

	list := svr.MembersByTags(map[string]string{"test-key": "test-value"})
	assert.Equal(t, 0, len(list))

	self := svr.LocalMember()
	assert.NotNil(t, self)

	m := svr.Member("syncer-test")
	assert.NotNil(t, m)

	cancel()
	svr.Stop()
}

func TestServerFailed(t *testing.T) {
	svr := NewServer(
		"",
		WithNode("syncer-test"),
		WithTags(map[string]string{"test-key": "test-value"}),
		WithAddTag("added-key", "added-value"),
		WithBindAddr("127.0.0.1"),
		WithBindPort(35151),
		WithAdvertiseAddr(""),
		WithAdvertisePort(0),
		WithEnableCompression(true),
		WithSecretKey([]byte("123456")),
		WithLogOutput(os.Stdout),
	)
	err := startServer(context.Background(), svr)
	assert.NotNil(t, err)

	svr.Stop()
}

func TestServerEventHandler(t *testing.T) {
	svr := defaultServer()
	startServer(context.Background(), svr)
	svr.OnceEventHandler(NewEventHandler(MemberJoinFilter(), func(data ...[]byte) bool {
		t.Log("Once event form member join triggered")
		return true
	}))

	// wait for trigger
	<-time.After(time.Second)

	handler := NewEventHandler(MemberJoinFilter(), func(data ...[]byte) bool {
		return false
	})

	svr.RemoveEventHandler(handler)

	svr.AddEventHandler(handler)

	svr.AddEventHandler(handler)

	svr.RemoveEventHandler(handler)

	svr.Stop()
}

func TestUserQuery(t *testing.T) {
	svr := defaultServer()
	startServer(context.Background(), svr)
	err := svr.Query("test-query", []byte("test-data"), func(from string, data []byte) {},
		WithFilterNodes("syncer-test"),
		WithFilterTags(map[string]string{"test-key": "test-value"}),
		WithRequestAck(false),
		WithRelayFactor(0),
		WithTimeout(time.Second),
	)
	assert.Nil(t, err)

	svr.Stop()
}

func TestEventHandler(t *testing.T) {
	filter := UserEventFilter("test-event")
	ok := filter.Invoke(serf.UserEvent{Name: "test-event", Payload: []byte("test-data")}, func(data ...[]byte) bool {
		return true
	})
	assert.True(t, ok)

	ok = filter.Invoke(&serf.Query{Name: "test-event", Payload: []byte("test-data")}, func(data ...[]byte) bool {
		return true
	})
	assert.False(t, ok)

	t.Log(filter.String())

	filter = QueryFilter("test-event")
	ok = filter.Invoke(&serf.Query{Name: "test-event", Payload: []byte("test-data")}, func(data ...[]byte) bool {
		return true
	})
	assert.True(t, ok)

	ok = filter.Invoke(serf.UserEvent{Name: "test-event", Payload: []byte("test-data")}, func(data ...[]byte) bool {
		return true
	})

	assert.False(t, ok)
	t.Log(filter.String())

	filter = MemberLeaveFilter()
	filter = MemberFailedFilter()
	filter = MemberUpdateFilter()
	filter = MemberReapFilter()
}

func defaultServer() *Server {
	return NewServer(
		"",
		WithNode("syncer-test"),
		WithBindAddr("127.0.0.1"),
		WithBindPort(35151),
	)
}

func startServer(ctx context.Context, svr *Server) (err error) {
	svr.Start(ctx)
	select {
	case <-svr.Ready():
	case <-svr.Stopped():
		err = errors.New("start serf server failed")
	case <-time.After(time.Second * 3):
		err = errors.New("start serf server timeout")
	}
	return
}
