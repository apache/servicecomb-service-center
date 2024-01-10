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

package pubsub

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"
	"golang.org/x/net/websocket"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/pubsub/ws"
)

var ErrRequiredServiceID = errors.New("required the serviceID")

// Watch listen the provider instance events by serviceID
func Watch(ctx context.Context, in *pb.WatchInstanceRequest, conn *websocket.Conn) {
	log.Info(fmt.Sprintf("new a web socket watch with service[%s]", in.SelfServiceId))
	if err := ExistService(ctx, in.SelfServiceId); err != nil {
		ws.SendEstablishError(conn, err)
		return
	}
	ws.Watch(ctx, in.SelfServiceId, conn)
}
func ExistService(ctx context.Context, selfServiceID string) error {
	if len(selfServiceID) == 0 {
		return ErrRequiredServiceID
	}
	resp, err := datasource.GetMetadataManager().ExistServiceByID(ctx, &pb.GetExistenceByIDRequest{
		ServiceId: selfServiceID,
	})
	if err != nil {
		log.Error("", err)
		return err
	}
	if !resp.Exist {
		return datasource.ErrServiceNotExists
	}
	return nil
}
