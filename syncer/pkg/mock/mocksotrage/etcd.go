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

package mocksotrage

import (
	"context"
	"errors"
	"os"

	"github.com/apache/servicecomb-service-center/syncer/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultName           = "etcd_mock"
	defaultDataDir        = "mock-data/"
	defaultListenPeerAddr = "127.0.0.1:30993"
)

type MockServer struct {
	etcd *etcd.Server
}

func NewKVServer() (svr *MockServer, err error) {
	agent, err1 := etcd.NewServer(defaultOptions()...)
	if err1 != nil {
		return nil, err
	}
	go agent.Start(context.Background())
	select {
	case <-agent.Ready():
	case <-agent.Stopped():
		return nil, errors.New("start etcd mock server failed")
	}
	return &MockServer{agent}, nil
}

func (m *MockServer) Storage() *clientv3.Client {
	return m.etcd.Storage()
}

func (m *MockServer) Stop() {
	m.etcd.Stop()
	os.RemoveAll(defaultDataDir)
}

func defaultOptions() []etcd.Option {
	return []etcd.Option{
		etcd.WithPeerAddr(defaultListenPeerAddr),
		etcd.WithName(defaultName),
		etcd.WithDataDir(defaultDataDir + defaultName),
	}
}
