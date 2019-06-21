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

func TestClient_RegisterInstance(t *testing.T) {
	svr, repo := newServiceCenter(t)
	_, err := repo.RegisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", &pb.SyncInstance{})
	if err != nil {
		t.Errorf("register instance failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.RegisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", &pb.SyncInstance{})
	if err != nil {
		t.Logf("register instance failed, error: %s", err)
	}
}

func TestClient_UnregisterInstance(t *testing.T) {
	svr, repo := newServiceCenter(t)
	err := repo.UnregisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Errorf("unregister instance failed, error: %s", err)
	}

	svr.Close()
	err = repo.UnregisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Logf("unregister instance failed, error: %s", err)
	}
}

func TestClient_Heartbeat(t *testing.T) {
	svr, repo := newServiceCenter(t)
	err := repo.Heartbeat(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Errorf("send instance heartbeat failed, error: %s", err)
	}

	svr.Close()
	err = repo.Heartbeat(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Logf("send instance heartbeat, error: %s", err)
	}
}
