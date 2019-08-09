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
	"fmt"
	"net/url"
	"os"

	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/coreos/etcd/clientv3"
)

const (
	defaultName           = "etcd_mock"
	defaultDataDir        = "mock-data/"
	defaultListenPeerAddr = "http://127.0.0.1:30993"
)

type MockServer struct {
	etcd *etcd.Agent
}

func NewKVServer() (svr *MockServer, err error) {
	agent := etcd.NewAgent(defaultConfig())
	go agent.Start(context.Background())
	select {
	case <-agent.Ready():
	case err = <-agent.Error():
	}
	if err != nil {
		return nil, err
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

func defaultConfig() *etcd.Config {
	peer, _ := url.Parse(defaultListenPeerAddr)
	conf := etcd.DefaultConfig()
	conf.Name = defaultName
	conf.Dir = defaultDataDir + defaultName
	conf.APUrls = []url.URL{*peer}
	conf.LPUrls = []url.URL{*peer}
	conf.InitialCluster = fmt.Sprintf("%s=%s", defaultName, defaultListenPeerAddr)
	return conf
}
