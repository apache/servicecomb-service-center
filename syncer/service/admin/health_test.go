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

package admin_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/apache/servicecomb-service-center/test"

	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/config"
	syncrpc "github.com/apache/servicecomb-service-center/syncer/rpc"
	"github.com/apache/servicecomb-service-center/syncer/service/admin"
	"github.com/stretchr/testify/assert"
)

type mockServer struct {
	v1sync.UnimplementedEventServiceServer
}

func (s *mockServer) Health(ctx context.Context, request *v1sync.HealthRequest) (*v1sync.HealthReply, error) {
	return &v1sync.HealthReply{Status: syncrpc.HealthStatusConnected}, nil
}

func TestHealth(t *testing.T) {
	port := 30101 + rand.Intn(1000)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	server := syncrpc.MockServer(addr, &mockServer{})
	defer server.Stop()

	c := config.GetConfig()
	tests := []struct {
		name    string
		sync    *config.Sync
		wantErr bool
	}{
		{name: "check no config ",
			sync:    nil,
			wantErr: true,
		},
		{name: "check disable is true",
			sync: &config.Sync{
				Peers: []*config.Peer{},
			},
			wantErr: true,
		},
		{name: "check no dataCenter",
			sync: &config.Sync{
				EnableOnStart: true,
				Peers:         []*config.Peer{},
			},
			wantErr: true,
		},
		{name: "check no endpoints",
			sync: &config.Sync{
				EnableOnStart: true,
				Peers: []*config.Peer{
					{Endpoints: nil},
				},
			},
			wantErr: true,
		},
		{name: "check endpoints is empty",
			sync: &config.Sync{
				EnableOnStart: true,
				Peers: []*config.Peer{
					{Endpoints: []string{}},
				},
			},
			wantErr: true,
		},

		{name: "given normal config",
			sync: &config.Sync{
				EnableOnStart: true,
				Peers: []*config.Peer{
					{Endpoints: []string{addr}},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		c.Sync = test.sync
		config.SetConfig(c)
		admin.Init()
		resp, err := admin.Health()
		hasErr := checkError(resp, err)
		assert.Equal(t, hasErr, test.wantErr, fmt.Sprintf("%s. health, wantErr %+v", test.name, test.wantErr))
	}
}

func checkError(resp *admin.Resp, err error) bool {
	if err != nil {
		return true
	}

	if resp.Peers == nil {
		return true
	}

	if len(resp.Peers) <= 0 {
		return true
	}
	return false
}

func TestHealthTotalTime(t *testing.T) {
	changeConfigPath()
	assert.NoError(t, config.Init())
	now := time.Now()
	_, err := admin.Health()
	assert.NoError(t, err)
	healthEndTime := time.Now()
	if healthEndTime.Sub(now) >= time.Second*30 {
		assert.NoError(t, errors.New("health api total time is too long"))
	}
}

func changeConfigPath() {
	workDir, _ := os.Getwd()
	replacePath := filepath.Join("syncer", "service", "admin")
	workDir = strings.ReplaceAll(workDir, replacePath, "etc")
	os.Setenv("APP_ROOT", workDir)
}
