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
	"net/url"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/model"

	"github.com/apache/servicecomb-service-center/pkg/log"
	scpb "github.com/apache/servicecomb-service-center/pkg/registry"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/gogo/protobuf/proto"
)

const (
	expansionDatasource = "datasource"
	expansionSchema     = "schema"
)

// toSyncData transform service-center service cache to SyncData
func toSyncData(cache *model.Cache, schemas []*scpb.Schema) (data *pb.SyncData) {
	data = &pb.SyncData{
		Services:  make([]*pb.SyncService, 0, len(cache.Microservices)),
		Instances: make([]*pb.SyncInstance, 0, len(cache.Instances)),
	}

	for _, service := range cache.Microservices {
		domain, project := getDomainProjectFromServiceKey(service.Key)
		if domain == "" {
			continue
		}
		syncService := toSyncService(service.Value)
		syncService.DomainProject = domain + "/" + project
		syncService.Expansions = append(syncService.Expansions, schemaExpansions(service.Value, schemas)...)

		syncInstances := toSyncInstances(syncService.ServiceId, cache.Instances)
		if len(syncInstances) == 0 {
			continue
		}

		data.Services = append(data.Services, syncService)
		data.Instances = append(data.Instances, syncInstances...)
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

//  toSyncInstances transform service-center instances to SyncInstances
func toSyncInstances(serviceID string, instances []*model.Instance) (syncInstances []*pb.SyncInstance) {
	for _, inst := range instances {
		if inst.Value.Status != scpb.MSI_UP {
			continue
		}

		if inst.Value.ServiceId == serviceID {
			syncInstances = append(syncInstances, toSyncInstance(serviceID, inst.Value))
		}
	}
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

// toService transform SyncService to service-center service
func toService(syncService *pb.SyncService) (service *scpb.MicroService) {
	service = &scpb.MicroService{}
	var err error
	if syncService.PluginName == PluginName && len(syncService.Expansions) > 0 {
		matches := pb.Expansions(syncService.Expansions).Find(expansionDatasource, map[string]string{})
		if len(matches) > 0 {
			err = proto.Unmarshal(matches[0].Bytes, service)
			if err == nil {
				service.ServiceId = syncService.ServiceId
				return
			}
			log.Errorf(err, "proto unmarshal %s service, serviceID = %s, kind = %v, content = %v failed",
				PluginName, service.ServiceId, matches[0].Kind, matches[0].Bytes)
		}
	}
	service.AppId = syncService.App
	service.ServiceId = syncService.ServiceId
	service.ServiceName = syncService.Name
	service.Version = syncService.Version
	service.Status = pb.SyncService_Status_name[int32(syncService.Status)]
	service.Environment = syncService.Environment
	return
}

// toInstance transform SyncInstance to service-center instance
func toInstance(syncInstance *pb.SyncInstance) (instance *scpb.MicroServiceInstance) {
	instance = &scpb.MicroServiceInstance{}
	if syncInstance.PluginName == PluginName && len(syncInstance.Expansions) > 0 {
		matches := pb.Expansions(syncInstance.Expansions).Find(expansionDatasource, map[string]string{})
		if len(matches) > 0 {
			err := proto.Unmarshal(matches[0].Bytes, instance)
			if err == nil {
				instance.InstanceId = syncInstance.InstanceId
				instance.ServiceId = syncInstance.ServiceId
				return
			}
			log.Errorf(err, "proto unmarshal %s instance, instanceID = %s, kind = %v, content = %v failed",
				PluginName, instance.InstanceId, matches[0].Kind, matches[0].Bytes)

		}
	}
	instance.InstanceId = syncInstance.InstanceId
	instance.ServiceId = syncInstance.ServiceId
	instance.Endpoints = make([]string, 0, len(syncInstance.Endpoints))
	instance.HostName = syncInstance.HostName
	instance.Version = syncInstance.Version
	instance.Status = pb.SyncInstance_Status_name[int32(syncInstance.Status)]

	for _, ep := range syncInstance.Endpoints {
		addr, err := url.Parse(ep)
		if err != nil {
			log.Errorf(err, "parse sc instance endpoint failed: %s", err)
			continue
		}
		endpoint := ""
		switch addr.Scheme {
		case "http":
			endpoint = strings.Replace(ep, "http://", "rest://", 1)
		case "https":
			endpoint = strings.Replace(ep, "https://", "rest://", 1) + "?sslEnabled=true"
		case "rest", "highway":
			endpoint = ep
		}
		instance.Endpoints = append(instance.Endpoints, endpoint)
	}

	if syncInstance.HealthCheck != nil && syncInstance.HealthCheck.Mode != pb.HealthCheck_UNKNOWN {
		instance.HealthCheck = &scpb.HealthCheck{
			Mode:     pb.HealthCheck_Modes_name[int32(syncInstance.HealthCheck.Mode)],
			Port:     syncInstance.HealthCheck.Port,
			Interval: syncInstance.HealthCheck.Interval,
			Times:    syncInstance.HealthCheck.Times,
			Url:      syncInstance.HealthCheck.Url,
		}
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
