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

package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/event"
	"github.com/apache/servicecomb-service-center/server/metrics"
	pb "github.com/go-chassis/cari/discovery"
)

type Broker struct {
	consumer *WebSocket
	producer *event.InstanceSubscriber
}

func (b *Broker) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case instanceEvent, ok := <-b.producer.Job:
			if !ok {
				return fmt.Errorf("read producer[%v] event failed", b.producer.Group())
			}
			err := b.write(instanceEvent)
			if err != nil {
				log.Errorf(err, "write instance event to subscriber[%s] failed, group: %s",
					b.consumer.RemoteAddr, b.producer.Group())
				return err
			}
		}
	}
}
func (b *Broker) write(evt *event.InstanceEvent) error {
	resp := evt.Response
	providerFlag := fmt.Sprintf("%s/%s/%s", resp.Key.AppId, resp.Key.ServiceName, resp.Key.Version)
	if resp.Action != string(pb.EVT_EXPIRE) {
		providerFlag = fmt.Sprintf("%s/%s(%s)", resp.Instance.ServiceId, resp.Instance.InstanceId, providerFlag)
	}
	remoteAddr := b.consumer.Conn.RemoteAddr().String()
	log.Infof("event[%s] is coming in, subscriber[%s] watch %s, group: %s",
		resp.Action, remoteAddr, providerFlag, b.producer.Group())

	resp.Response = nil
	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorf(err, "subscriber[%s] watch %s, group: %s",
			remoteAddr, providerFlag, b.producer.Group())
		data = util.StringToBytesWithNoCopy(fmt.Sprintf("marshal output file error, %s", err.Error()))
	}
	err = b.consumer.WriteTextMessage(data)
	metrics.ReportPublishCompleted(evt, err)
	return err
}

func NewBroker(ws *WebSocket, is *event.InstanceSubscriber) *Broker {
	return &Broker{
		consumer: ws,
		producer: is,
	}
}
