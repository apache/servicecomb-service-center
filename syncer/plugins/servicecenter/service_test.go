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
package servicecenter

import (
	"context"
	"testing"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

func TestClient_CreateService(t *testing.T) {
	svr, repo := newServiceCenter(t)
	_, err := repo.CreateService(context.Background(), "default/deault", &pb.SyncService{})
	if err != nil {
		t.Errorf("create service failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.CreateService(context.Background(), "default/deault", &pb.SyncService{})
	if err != nil {
		t.Logf("create service failed, error: %s", err)
	}
}

func TestClient_ServiceExistence(t *testing.T) {
	svr, repo := newServiceCenter(t)
	_, err := repo.ServiceExistence(context.Background(), "default/deault", &pb.SyncService{})
	if err != nil {
		t.Errorf("check service existence failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.ServiceExistence(context.Background(), "default/deault", &pb.SyncService{})
	if err != nil {
		t.Logf("check service existence failed, error: %s", err)
	}
}

func TestClient_DeleteService(t *testing.T) {
	svr, repo := newServiceCenter(t)
	err := repo.DeleteService(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd")
	if err != nil {
		t.Errorf("delete service failed, error: %s", err)
	}

	svr.Close()
	err = repo.DeleteService(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd")
	if err != nil {
		t.Logf("delete service failed, error: %s", err)
	}
}
