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

package heartbeat

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/gorilla/websocket"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/pubsub/ws"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
)

const (
	Websocket         = "Websocket"
	defaultPingPeriod = 30 * time.Second
	minPeriod         = 1 * time.Second
	maxPeriod         = 1 * time.Hour
)

var (
	once       sync.Once
	pingPeriod time.Duration
)

type client struct {
	cxt        context.Context
	conn       *websocket.Conn
	serviceID  string
	instanceID string
}

func configuration() {
	once.Do(func() {
		pingPeriod = config.GetDuration("heartbeat.websocket.pingInterval", defaultPingPeriod)
		if pingPeriod < minPeriod || pingPeriod > maxPeriod {
			pingPeriod = defaultPingPeriod
		}
	})
}

func newClient(ctx context.Context, conn *websocket.Conn, serviceID string, instanceID string) *client {
	configuration()
	return &client{
		cxt:        ctx,
		conn:       conn,
		serviceID:  serviceID,
		instanceID: instanceID,
	}
}

func (c *client) sendClose(code int, text string) error {
	remoteAddr := c.conn.RemoteAddr().String()
	var message []byte
	if code != websocket.CloseNoStatusReceived {
		message = websocket.FormatCloseMessage(code, text)
	}
	err := c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(ws.SendTimeout))
	if err != nil {
		log.Error(fmt.Sprintf("watcher[%s] catch an err", remoteAddr), err)
		return err
	}
	return nil
}

func (c *client) heartbeat() {
	remoteAddr := c.conn.RemoteAddr().String()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		<-ticker.C
		err := c.conn.SetWriteDeadline(time.Now().Add(ws.SendTimeout))
		if err != nil {
			log.Error("", err)
		}
		if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			log.Error(fmt.Sprintf("send 'Ping' message to watcher[%s] failed", remoteAddr), err)
			return
		}
	}
}

func (c *client) handleMessage() {
	defer func() {
		c.conn.Close()
	}()

	remoteAddr := c.conn.RemoteAddr().String()
	c.conn.SetPongHandler(func(message string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(ws.ReadTimeout))
		if err != nil {
			log.Error("", err)
		}
		log.Infof("received 'Pong' message '%s' from watcher[%s]\n", message, remoteAddr)
		request := &pb.HeartbeatRequest{
			ServiceId:  c.serviceID,
			InstanceId: c.instanceID,
		}
		_, err = discosvc.Heartbeat(c.cxt, request)
		if err != nil {
			log.Error("instance heartbeat report failed ", err)
		}
		return err
	})

	c.conn.SetCloseHandler(func(code int, text string) error {
		log.Info(fmt.Sprintf("watcher[%s] active closed, code: %d, message: '%s'", remoteAddr, code, text))
		return c.sendClose(code, text)
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("", err)
			}
			break
		}
	}
}

func SendEstablishError(conn *websocket.Conn, err error) {
	remoteAddr := conn.RemoteAddr().String()
	log.Errorf(err, "establish[%s] websocket failed.", remoteAddr)
	if err := conn.WriteMessage(websocket.TextMessage, util.StringToBytesWithNoCopy(err.Error())); err != nil {
		log.Errorf(err, "establish[%s] websocket failed: write message failed.", remoteAddr)
	}
}

func Heartbeat(ctx context.Context, conn *websocket.Conn, serviceID string, instanceID string) {
	domain := util.ParseDomain(ctx)
	client := newClient(ctx, conn, serviceID, instanceID)
	metrics.ReportSubscriber(domain, Websocket, 1)
	process(client)
	metrics.ReportSubscriber(domain, Websocket, -1)
}

func process(client *client) {
	go client.heartbeat()
	client.handleMessage()
}

func WatchHeartbeat(ctx context.Context, in *pb.HeartbeatRequest, conn *websocket.Conn) {
	log.Info(fmt.Sprintf("new a web socket with service[%s] ,instance[%s]", in.ServiceId, in.InstanceId))
	if err := preOp(ctx, in); err != nil {
		SendEstablishError(conn, err)
		return
	}
	Heartbeat(ctx, conn, in.ServiceId, in.InstanceId)
}
func preOp(ctx context.Context, in *pb.HeartbeatRequest) error {
	if in == nil || len(in.ServiceId) == 0 || len(in.InstanceId) == 0 {
		return errors.New("request format invalid")
	}
	resp, err := datasource.GetMetadataManager().ExistInstanceByID(ctx, &pb.MicroServiceInstanceKey{
		ServiceId:  in.ServiceId,
		InstanceId: in.InstanceId,
	})
	if err != nil {
		return err
	}
	if !resp.Exist {
		return datasource.ErrInstanceNotExists
	}
	return nil
}
