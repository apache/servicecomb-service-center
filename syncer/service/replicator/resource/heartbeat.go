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

package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/cari/sync"
)

const (
	Heartbeat = "heartbeat"
)

func NewHeartbeat(e *v1sync.Event) Resource {
	h := &heartbeat{
		event: e,
	}
	h.manager = new(metadataManage)
	return h
}

type heartbeat struct {
	event *v1sync.Event

	input *pb.HeartbeatRequest

	manager metadataManager
	defaultFailHandler
}

func (h *heartbeat) loadInput() error {
	h.input = new(pb.HeartbeatRequest)

	return json.Unmarshal(h.event.Value, h.input)
}

func (h *heartbeat) LoadCurrentResource(_ context.Context) *Result {
	err := h.loadInput()
	if err != nil {
		return FailResult(err)
	}

	return nil
}

func (h *heartbeat) NeedOperate(context.Context) *Result {
	return nil
}

func MicroNonExistResult() *Result {
	return NewResult(MicroNonExist, "")
}

func InstNonExistResult() *Result {
	return NewResult(InstNonExist, "")
}

func (h *heartbeat) Operate(ctx context.Context) *Result {
	err := h.manager.SendHeartbeat(ctx, h.input)
	if err == nil {
		return SuccessResult()
	}

	log.Warn(fmt.Sprintf("send heartbeat failed, %s, %s",
		h.input.ServiceId, h.input.InstanceId))

	_, err = h.manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
		ProviderServiceId:  h.input.ServiceId,
		ProviderInstanceId: h.input.InstanceId,
	})

	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists) {
			return MicroNonExistResult()
		}

		if errsvc.IsErrEqualCode(err, pb.ErrInstanceNotExists) {
			return InstNonExistResult()
		}
		return FailResult(err)
	}

	log.Info(fmt.Sprintf("get instance return exist, %s, %s",
		h.input.ServiceId, h.input.InstanceId))
	return SuccessResult()
}

func (h *heartbeat) FailHandle(ctx context.Context, code int32) (*v1sync.Event, error) {
	if code != InstNonExist {
		return nil, nil
	}

	err := h.loadInput()
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("instance %s,%s not exist, start rebuild instance",
		h.input.ServiceId, h.input.InstanceId))

	return h.rebuildInstance(ctx)
}

func (h *heartbeat) rebuildInstance(ctx context.Context) (*v1sync.Event, error) {
	ctx = util.SetDomain(ctx, h.event.Opts[string(util.CtxDomain)])
	ctx = util.SetProject(ctx, h.event.Opts[string(util.CtxProject)])

	cur, err := h.manager.GetInstance(ctx, &pb.GetOneInstanceRequest{
		ProviderServiceId:  h.input.ServiceId,
		ProviderInstanceId: h.input.InstanceId,
	})
	if err != nil {
		if errsvc.IsErrEqualCode(err, pb.ErrInstanceNotExists) {
			log.Warn(fmt.Sprintf("instance %s,%s not exist",
				h.input.ServiceId, h.input.InstanceId))
			return nil, nil
		}
		return nil, err
	}

	value, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}

	eventID, err := v1sync.NewEventID()
	if err != nil {
		log.Error("fail to create eventID", err)
		return nil, err
	}

	return &v1sync.Event{
		Id:        eventID,
		Action:    sync.CreateAction,
		Subject:   Instance,
		Opts:      h.event.Opts,
		Value:     value,
		Timestamp: v1sync.Timestamp(),
	}, nil
}
