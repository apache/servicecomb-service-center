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
package event

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/discovery"
	"github.com/apache/servicecomb-service-center/server/service/cache"
	"github.com/apache/servicecomb-service-center/server/service/metrics"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
	"strings"
)

const (
	increaseOne = 1
	decreaseOne = -1
)

// InstanceEventHandler is the handler to handle:
// 1. report instance metrics
// 2. recover the instance quota
// 3. publish the instance events to the subscribers
// 4. reset the find instance cache
type InstanceEventHandler struct {
}

func (h *InstanceEventHandler) Type() discovery.Type {
	return backend.INSTANCE
}

func (h *InstanceEventHandler) OnEvent(evt discovery.KvEvent) {
	action := evt.Type
	instance := evt.KV.Value.(*pb.MicroServiceInstance)
	providerId, providerInstanceId, domainProject := apt.GetInfoFromInstKV(evt.KV.Key)
	idx := strings.Index(domainProject, "/")
	domainName := domainProject[:idx]
	projectName := domainProject[idx+1:]

	var count float64 = increaseOne
	switch action {
	case pb.EVT_INIT:
		metrics.ReportInstances(domainName, count)
		ms := serviceUtil.GetServiceFromCache(domainProject, providerId)
		if ms == nil {
			log.Warnf("caught [%s] instance[%s/%s] event, endpoints %v, get cached provider's file failed",
				action, providerId, providerInstanceId, instance.Endpoints)
			return
		}
		frameworkName, frameworkVersion := getFramework(ms)
		metrics.ReportFramework(domainName, projectName, frameworkName, frameworkVersion, count)
		return
	case pb.EVT_CREATE:
		metrics.ReportInstances(domainName, count)
	case pb.EVT_DELETE:
		count = decreaseOne
		metrics.ReportInstances(domainName, count)
		if !apt.IsDefaultDomainProject(domainProject) {
			projectName := domainProject[idx+1:]
			serviceUtil.RemandInstanceQuota(
				util.SetDomainProject(context.Background(), domainName, projectName))
		}
	}

	if notify.NotifyCenter().Closed() {
		log.Warnf("caught [%s] instance[%s/%s] event, endpoints %v, but notify service is closed",
			action, providerId, providerInstanceId, instance.Endpoints)
		return
	}

	// 查询服务版本信息
	ctx := context.WithValue(context.WithValue(context.Background(),
		serviceUtil.CTX_CACHEONLY, "1"),
		serviceUtil.CTX_GLOBAL, "1")
	ms, err := serviceUtil.GetService(ctx, domainProject, providerId)
	if ms == nil {
		log.Errorf(err, "caught [%s] instance[%s/%s] event, endpoints %v, get cached provider's file failed",
			action, providerId, providerInstanceId, instance.Endpoints)
		return
	}

	frameworkName, frameworkVersion := getFramework(ms)
	metrics.ReportFramework(domainName, projectName, frameworkName, frameworkVersion, count)

	log.Infof("caught [%s] service[%s][%s/%s/%s/%s] instance[%s] event, endpoints %v",
		action, providerId, ms.Environment, ms.AppId, ms.ServiceName, ms.Version,
		providerInstanceId, instance.Endpoints)

	// 查询所有consumer
	consumerIds, _, err := serviceUtil.GetAllConsumerIds(ctx, domainProject, ms)
	if err != nil {
		log.Errorf(err, "get service[%s][%s/%s/%s/%s]'s consumerIds failed",
			providerId, ms.Environment, ms.AppId, ms.ServiceName, ms.Version)
		return
	}

	PublishInstanceEvent(evt, domainProject, pb.MicroServiceToKey(domainProject, ms), consumerIds)
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func PublishInstanceEvent(evt discovery.KvEvent, domainProject string, serviceKey *pb.MicroServiceKey, subscribers []string) {
	defer cache.FindInstances.Remove(serviceKey)

	if len(subscribers) == 0 {
		return
	}

	response := &pb.WatchInstanceResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Watch instance successfully."),
		Action:   string(evt.Type),
		Key:      serviceKey,
		Instance: evt.KV.Value.(*pb.MicroServiceInstance),
	}
	for _, consumerId := range subscribers {
		// TODO add超时怎么处理？
		job := notify.NewInstanceEventWithTime(consumerId, domainProject, evt.Revision, evt.CreateAt, response)
		err := notify.NotifyCenter().Publish(job)
		if err != nil {
			log.Errorf(err, "publish job failed")
		}
	}
}
