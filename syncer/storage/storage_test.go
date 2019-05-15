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

package storage

import (
	"os"
	"path/filepath"
	"testing"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

func TestSyncData(t *testing.T) {
	storage := New()
	defer func() {
		storage.Stop()
		os.RemoveAll(filepath.Dir(snapshotPath))
	}()
	data := storage.GetSyncData()
	if len(data.Services) > 0 {
		t.Error("default sync data was wrong")
	}
	data = getSyncData()
	storage.SaveSyncData(data)
	nd := storage.GetSyncData()
	if nd == nil {
		t.Error("save sync data failed!")
	}
	storage.Stop()

	New()
}

func TestSyncMapping(t *testing.T) {
	storage := New()
	defer func() {
		storage.Stop()
		os.RemoveAll(filepath.Dir(snapshotPath))
	}()
	nodeName := "testnode"
	data := storage.GetSyncMapping(nodeName)
	if len(data) > 0 {
		t.Error("default sync mapping was wrong")
	}
	data = getSyncMapping(nodeName)
	storage.SaveSyncMapping(nodeName, data)
	nd := storage.GetSyncMapping(nodeName)
	_, ok := nd[nodeName]
	if !ok {
		t.Error("save sync mapping failed!")
	}

	all := storage.GetAllMapping()
	for key := range all {
		if key == nodeName {
			return
		}
	}

	t.Errorf("all mapping has not node name %s!", nodeName)
}

func getSyncMapping(nodeName string) pb.SyncMapping {
	return pb.SyncMapping{nodeName: &pb.SyncServiceKey{
		DomainProject: "default/default",
		ServiceID:     "5db1b794aa6f8a875d6e68110260b5491ee7e223",
		InstanceID:    "4d41a637471f11e9888cfa163eca30e0",
	}}
}

func getSyncData() *pb.SyncData {
	return &pb.SyncData{
		Services: []*pb.SyncService{
			{
				DomainProject: "default/default",
				Service: &scpb.MicroService{
					ServiceId:   "5db1b794aa6f8a875d6e68110260b5491ee7e223",
					AppId:       "default",
					ServiceName: "SERVICECENTER",
					Version:     "1.1.0",
					Level:       "BACK",
					Schemas: []string{
						"servicecenter.grpc.api.ServiceCtrl",
						"servicecenter.grpc.api.ServiceInstanceCtrl",
					},
					Status: "UP",
					Properties: map[string]string{
						"allowCrossApp": "true",
					},
					Timestamp:    "1552626180",
					ModTimestamp: "1552626180",
					Environment:  "production",
				},
				Instances: []*scpb.MicroServiceInstance{
					{
						InstanceId: "4d41a637471f11e9888cfa163eca30e0",
						ServiceId:  "5db1b794aa6f8a875d6e68110260b5491ee7e223",
						Endpoints: []string{
							"rest://127.0.0.1:30100/",
						},
						HostName: "testmock",
						Status:   "UP",
						HealthCheck: &scpb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    3,
						},
						Timestamp:    "1552653537",
						ModTimestamp: "1552653537",
						Version:      "1.1.0",
					},
				},
			},
		},
	}
}
