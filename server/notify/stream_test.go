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
package notify

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/log"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type grpcWatchServer struct {
	grpc.ServerStream
}

func (x *grpcWatchServer) Send(m *pb.WatchInstanceResponse) error {
	return nil
}

func (x *grpcWatchServer) Context() context.Context {
	return context.Background()
}

func TestHandleWatchJob(t *testing.T) {
	w := NewInstanceEventListWatcher("g", "s", nil)
	w.Job <- nil
	err := HandleWatchJob(w, &grpcWatchServer{})
	if err == nil {
		t.Fatalf("TestHandleWatchJob failed")
	}
	w.Job <- NewInstanceEventWithTime("g", "s", 1, simple.FromTime(time.Now()), nil)
	w.Job <- nil
	HandleWatchJob(w, &grpcWatchServer{})
}

func TestDoStreamListAndWatch(t *testing.T) {
	defer log.Recover()
	err := DoStreamListAndWatch(context.Background(), "s", nil, nil)
	t.Fatal("TestDoStreamListAndWatch failed", err)
}
