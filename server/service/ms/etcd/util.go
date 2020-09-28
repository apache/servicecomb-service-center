// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/uuid"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"strconv"
	"time"
)

func GetSingleService(ctx context.Context, domainProject string, serviceID string) (*pb.MicroService, error) {
	key := apt.GenerateServiceKey(domainProject, serviceID)
	opts := append(serviceUtil.FromContext(ctx), registry.WithStrKey(key))
	serviceResp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(serviceResp.Kvs) == 0 {
		return nil, nil
	}
	return serviceResp.Kvs[0].Value.(*pb.MicroService), nil
}

func GetAllServiceUtil(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	services, err := GetServicesByDomainProject(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return services, nil
}

func GetServicesByDomainProject(ctx context.Context, domainProject string) ([]*pb.MicroService, error) {
	kvs, err := getServicesRawData(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	var services []*pb.MicroService
	for _, kv := range kvs {
		services = append(services, kv.Value.(*pb.MicroService))
	}
	return services, nil
}

func getServicesRawData(ctx context.Context, domainProject string) ([]*discovery.KeyValue, error) {
	key := apt.GenerateServiceKey(domainProject, "")
	opts := append(serviceUtil.FromContext(ctx),
		registry.WithStrKey(key),
		registry.WithPrefix())
	resp, err := backend.Store().Service().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, err
}

func CapRegisterData(ctx context.Context, request *pb.CreateServiceRequest) (
	[]registry.PluginOp, []registry.CompareOp, []registry.PluginOp, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceBody := request.Service
	serviceFlag := util.StringJoin([]string{
		serviceBody.Environment, serviceBody.AppId, serviceBody.ServiceName, serviceBody.Version}, "/")
	serviceUtil.SetServiceDefaultValue(serviceBody)
	domainProject := util.ParseDomainProject(ctx)
	serviceKey := &pb.MicroServiceKey{
		Tenant:      domainProject,
		Environment: serviceBody.Environment,
		AppId:       serviceBody.AppId,
		ServiceName: serviceBody.ServiceName,
		Alias:       serviceBody.Alias,
		Version:     serviceBody.Version,
	}
	index := apt.GenerateServiceIndexKey(serviceKey)

	// generate global service id
	requestServiceID := serviceBody.ServiceId
	if len(requestServiceID) == 0 {
		ctx = util.SetContext(ctx, uuid.ContextKey, index)
		serviceBody.ServiceId = plugin.Plugins().UUID().GetServiceID(ctx)
	}
	serviceBody.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	serviceBody.ModTimestamp = serviceBody.Timestamp

	data, err := json.Marshal(serviceBody)
	if err != nil {
		log.Errorf(err, "create micro-serviceBody[%s] failed, json marshal serviceBody failed, operator: %s",
			serviceFlag, remoteIP)
		return nil, nil, nil, err
	}

	key := apt.GenerateServiceKey(domainProject, serviceBody.ServiceId)
	keyBytes := util.StringToBytesWithNoCopy(key)
	indexBytes := util.StringToBytesWithNoCopy(index)
	aliasBytes := util.StringToBytesWithNoCopy(apt.GenerateServiceAliasKey(serviceKey))

	opts := []registry.PluginOp{
		registry.OpPut(registry.WithKey(keyBytes), registry.WithValue(data)),
		registry.OpPut(registry.WithKey(indexBytes), registry.WithStrValue(serviceBody.ServiceId)),
	}
	uniqueCmpOpts := []registry.CompareOp{
		registry.OpCmp(registry.CmpVer(indexBytes), registry.CmpEqual, 0),
		registry.OpCmp(registry.CmpVer(keyBytes), registry.CmpEqual, 0),
	}
	failOpts := []registry.PluginOp{
		registry.OpGet(registry.WithKey(indexBytes)),
	}

	if len(serviceKey.Alias) > 0 {
		opts = append(opts, registry.OpPut(registry.WithKey(aliasBytes), registry.WithStrValue(serviceBody.ServiceId)))
		uniqueCmpOpts = append(uniqueCmpOpts,
			registry.OpCmp(registry.CmpVer(aliasBytes), registry.CmpEqual, 0))
		failOpts = append(failOpts, registry.OpGet(registry.WithKey(aliasBytes)))
	}

	return opts, uniqueCmpOpts, failOpts, nil
}

func NewRegisterServiceResp(ctx context.Context, reqService *pb.MicroService, resp *registry.PluginResponse,
	err error) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)
	serviceFlag := util.StringJoin([]string{
		reqService.Environment, reqService.AppId, reqService.ServiceName, reqService.Version}, "/")
	if err != nil {
		log.Errorf(err, "create micro-service[%s] failed, operator: %s",
			serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response: proto.CreateResponse(scerr.ErrUnavailableBackend, err.Error()),
		}, err
	}

	if !resp.Succeeded {
		requestServiceID := reqService.ServiceId
		if len(requestServiceID) != 0 {
			if len(resp.Kvs) == 0 ||
				requestServiceID != util.BytesToStringWithNoCopy(resp.Kvs[0].Value) {
				log.Warnf("create micro-service[%s] failed, service already exists, operator: %s",
					serviceFlag, remoteIP)
				return &pb.CreateServiceResponse{
					Response: proto.CreateResponse(scerr.ErrServiceAlreadyExists,
						"ServiceID conflict or found the same service with different id."),
				}, nil
			}
		}

		if len(resp.Kvs) == 0 {
			// internal error?
			log.Errorf(nil, "create micro-service[%s] failed, unexpected txn response, operator: %s",
				serviceFlag, remoteIP)
			return &pb.CreateServiceResponse{
				Response: proto.CreateResponse(scerr.ErrInternal, "Unexpected txn response."),
			}, nil
		}

		serviceIDInner := util.BytesToStringWithNoCopy(resp.Kvs[0].Value)
		log.Warnf("create micro-service[%s][%s] failed, service already exists, operator: %s",
			serviceIDInner, serviceFlag, remoteIP)
		return &pb.CreateServiceResponse{
			Response:  proto.CreateResponse(proto.Response_SUCCESS, "register service successfully"),
			ServiceId: serviceIDInner,
		}, nil
	}

	log.Infof("create micro-service[%s][%s] successfully, operator: %s",
		reqService.ServiceId, serviceFlag, remoteIP)
	return &pb.CreateServiceResponse{
		Response:  proto.CreateResponse(proto.Response_SUCCESS, "Register service successfully."),
		ServiceId: reqService.ServiceId,
	}, nil
}
