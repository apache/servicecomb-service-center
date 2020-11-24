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
package server

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	sc "github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	scpb "github.com/go-chassis/cari/discovery"
	"github.com/gogo/protobuf/proto"
)

const (
	expansionDatasource = "datasource"
	expansionSchema     = "schema"
	expansionAction     = "action"
	PluginName          = "servicecenter"
)

func (s *Server) EventQueueToSyncData(ctx context.Context, incrementQueue []*dump.WatchInstanceChangedEvent) (data *pb.SyncData) {
	data = &pb.SyncData{
		Services:  make([]*pb.SyncService, 0, len(incrementQueue)),
		Instances: make([]*pb.SyncInstance, 0, len(incrementQueue)),
	}
	cli, err := sc.NewSCClient(plugins.ToSCConfig(convertSCConfigOption(s.conf)...))
	if err != nil {
		log.Error("create scClient failed: %s", err)
	}
	for _, event := range incrementQueue {
		service := event.Service
		instance := event.Instance
		domain, project := getDomainProjectFromServiceKey(service.Key)
		if domain == "" {
			continue
		}

		syncService := toSyncService(service.Value)
		syncService.DomainProject = domain + "/" + project

		ss, err := cli.GetSchemasByServiceID(ctx, domain, project, service.Value.ServiceId)
		if err != nil {
			log.Warnf("get schemas by serviceId failed: %s", err)
		}
		syncService.Expansions = append(syncService.Expansions, schemaExpansions(service.Value, ss)...)

		syncInstance := toSyncInstance(syncService.ServiceId, instance.Value)
		syncInstance.Expansions = append(syncInstance.Expansions, actionExpansions(event)...)

		data.Services = append(data.Services, syncService)
		data.Instances = append(data.Instances, syncInstance)
	}
	return
}

// toSyncService transform service-center service to SyncService
func toSyncService(service *scpb.MicroService) (syncService *pb.SyncService) {
	syncService = &pb.SyncService{
		ServiceId:   service.ServiceId,
		App:         service.AppId,
		Name:        service.ServiceName,
		Version:     service.Version,
		Environment: service.Environment,
		PluginName:  PluginName,
	}
	switch service.Status {
	case scpb.MS_UP:
		syncService.Status = pb.SyncService_UP
	case scpb.MS_DOWN:
		syncService.Status = pb.SyncService_DOWN
	default:
		syncService.Status = pb.SyncService_UNKNOWN
	}

	content, err := proto.Marshal(service)
	if err != nil {
		log.Errorf(err, "transform sc service to syncer service failed: %s", err)
		return
	}

	syncService.Expansions = []*pb.Expansion{{
		Kind:   expansionDatasource,
		Bytes:  content,
		Labels: map[string]string{},
	}}
	return
}

// toSyncInstance transform service-center instance to SyncInstance
func toSyncInstance(serviceID string, instance *scpb.MicroServiceInstance) (syncInstance *pb.SyncInstance) {
	syncInstance = &pb.SyncInstance{
		InstanceId: instance.InstanceId,
		ServiceId:  serviceID,
		Endpoints:  make([]string, 0, len(instance.Endpoints)),
		HostName:   instance.HostName,
		Version:    instance.Version,
		PluginName: PluginName,
	}
	switch instance.Status {
	case scpb.MSI_UP:
		syncInstance.Status = pb.SyncInstance_UP
	case scpb.MSI_DOWN:
		syncInstance.Status = pb.SyncInstance_DOWN
	case scpb.MSI_STARTING:
		syncInstance.Status = pb.SyncInstance_STARTING
	case scpb.MSI_OUTOFSERVICE:
		syncInstance.Status = pb.SyncInstance_OUTOFSERVICE
	default:
		syncInstance.Status = pb.SyncInstance_UNKNOWN
	}

	for _, ep := range instance.Endpoints {
		endpoint := ep
		addr, err := url.Parse(ep)
		if err != nil {
			log.Errorf(err, "parse sc instance endpoint failed: %s", err)
			continue
		}
		if addr.Scheme == "rest" {
			prefix := "http://"
			b, _ := strconv.ParseBool(addr.Query().Get("sslEnabled"))
			if b {
				prefix = "https://"
			}
			endpoint = strings.Replace(ep, addr.Scheme+"://", prefix, 1)
		}
		syncInstance.Endpoints = append(syncInstance.Endpoints, endpoint)
	}

	if instance.HealthCheck != nil {
		syncInstance.HealthCheck = &pb.HealthCheck{
			Port:     instance.HealthCheck.Port,
			Interval: instance.HealthCheck.Interval,
			Times:    instance.HealthCheck.Times,
			Url:      instance.HealthCheck.Url,
		}
	}

	content, err := proto.Marshal(instance)
	if err != nil {
		log.Errorf(err, "transform sc instance to syncer instance failed: %s", err)
		return
	}

	syncInstance.Expansions = []*pb.Expansion{{
		Kind:   expansionDatasource,
		Bytes:  content,
		Labels: map[string]string{},
	}}
	return
}

func schemaExpansions(service *scpb.MicroService, schemas []*scpb.Schema) (expansions []*pb.Expansion) {
	for _, val := range schemas {
		if !inSlice(service.Schemas, val.SchemaId) {
			continue
		}

		content, err := proto.Marshal(val)
		if err != nil {
			log.Errorf(err, "proto marshal schemas failed, app = %s, service = %s, version = %s datasource = %s",
				service.AppId, service.ServiceName, service.Version, expansionSchema)
			continue
		}
		expansions = append(expansions, &pb.Expansion{
			Kind:   expansionSchema,
			Bytes:  content,
			Labels: map[string]string{},
		})
	}
	return
}

func getDomainProjectFromServiceKey(serviceKey string) (string, string) {
	tenant := strings.Split(serviceKey, "/")
	if len(tenant) < 6 {
		return "", ""
	}
	return tenant[4], tenant[5]
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func actionExpansions(event *dump.WatchInstanceChangedEvent) (expansions []*pb.Expansion) {
	action := event.Action
	content := []byte(action)
	expansions = append(expansions, &pb.Expansion{
		Kind:   expansionAction,
		Bytes:  content,
		Labels: map[string]string{},
	})
	return
}
