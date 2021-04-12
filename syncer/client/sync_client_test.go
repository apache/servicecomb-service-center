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
	"context"
	"crypto/tls"
	"testing"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/stretchr/testify/assert"
)

var c = NewSyncClient("", new(tls.Config))

func TestClient_IncrementPull(t *testing.T) {
	t.Run("Test IncrementPull", func(t *testing.T) {
		_, err := c.IncrementPull(context.Background(), &pb.IncrementPullRequest{Addr: "http://127.0.0.1", Length: 3})
		assert.Error(t, err, "IncrementPull fail without grpc")
	})
}

func TestClient_DeclareDataLength(t *testing.T) {
	t.Run("DeclareDataLength test", func(t *testing.T) {
		_, err := c.DeclareDataLength(context.Background(), "http://127.0.0.1")
		assert.Error(t, err, "DeclareDataLength fail without grpc")
	})
}

func syncDataCreate() *pb.SyncData {
	syncService := pb.SyncService{
		ServiceId: "a59f99611a6945677a21f28c0aeb05abb",
	}
	services := []*pb.SyncService{&syncService}
	syncInstance := pb.SyncInstance{
		InstanceId: "5e1140fc232111eb9bb600acc8c56b5b",
	}
	instances := []*pb.SyncInstance{&syncInstance}
	syncData := pb.SyncData{
		Services:  services,
		Instances: instances,
	}
	return &syncData
}

func declareRespCreate() *pb.DeclareResponse {
	declareResp := pb.DeclareResponse{
		SyncDataLength: 3,
	}
	return &declareResp
}
