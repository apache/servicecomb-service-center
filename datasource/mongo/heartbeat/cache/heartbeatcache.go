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

package heartbeatcache

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource/mongo/heartbeat"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	maxInterval     = 60
	defaultInterval = 30
	maxTimes        = 3
	minTimes        = 0
)

var ErrHeartbeatConversionFailed = errors.New("instanceHeartbeatInfo type conversion failed. ")

func init() {
	heartbeat.Install("cache", NewHeartBeatCache)
}

type HeartBeatCache struct {
	Cfg *CacheConfig
}

func NewHeartBeatCache() (heartbeat.HealthCheck, error) {
	return &HeartBeatCache{Cfg: Configuration()}, nil
}

func (h *HeartBeatCache) Heartbeat(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if ins, ok := h.Cfg.InstanceHeartbeatStore.Get(request.InstanceId); ok {
		return h.inCacheStrategy(ctx, request, ins)
	}
	return h.notInCacheStrategy(ctx, request)
}

// CheckInstance func is to add instance related information to the cache
func (h *HeartBeatCache) CheckInstance(ctx context.Context, instance *pb.MicroServiceInstance) error {
	return h.Cfg.AddHeartbeatTask(instance.ServiceId, instance.InstanceId, instance.HealthCheck.Interval*(instance.HealthCheck.Times+1))
}

func (h *HeartBeatCache) inCacheStrategy(ctx context.Context, request *pb.HeartbeatRequest, insHeartbeatInfo interface{}) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	heartbeatInfo, ok := insHeartbeatInfo.(*InstanceHeartbeatInfo)
	if !ok {
		log.Error("type conversion failed: %v", ErrHeartbeatConversionFailed)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrInstanceNotExists, ErrHeartbeatConversionFailed.Error())),
		}
		return resp, ErrHeartbeatConversionFailed
	}
	err := h.Cfg.AddHeartbeatTask(request.ServiceId, request.InstanceId, heartbeatInfo.TTL)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrNotEnoughQuota, err.Error())),
		}
		return resp, err
	}
	err = updateInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrInstanceNotExists, err.Error())),
		}
		return resp, err
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "update service instance heartbeat successfully"),
	}, nil
}

func (h *HeartBeatCache) notInCacheStrategy(ctx context.Context, request *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	instance, err := findInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrInstanceNotExists, err.Error())),
		}
		return resp, err
	}
	interval, times := instance.Instance.HealthCheck.Interval, instance.Instance.HealthCheck.Times
	// Set the range of interval and time
	if interval > maxInterval || interval < minTimes {
		interval = defaultInterval
	}
	if times > maxTimes || times < minTimes {
		times = maxTimes
	}
	err = h.Cfg.AddHeartbeatTask(request.ServiceId, request.InstanceId, interval*(times+1))
	if err != nil {
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrNotEnoughQuota, err.Error())),
		}
		return resp, err
	}
	err = updateInstance(ctx, request.ServiceId, request.InstanceId)
	if err != nil {
		h.Cfg.RemoveCacheInstance(request.InstanceId)
		log.Error(fmt.Sprintf("heartbeat failed, instance[%s]. operator %s", request.InstanceId, remoteIP), err)
		resp := &pb.HeartbeatResponse{
			Response: pb.CreateResponseWithSCErr(pb.NewError(pb.ErrInstanceNotExists, err.Error())),
		}
		return resp, err
	}
	return &pb.HeartbeatResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "update service instance heartbeat successfully"),
	}, nil
}
