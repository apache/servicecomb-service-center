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

package eureka

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const (
	eurekaInstanceFormat = "%s:%s:%d"

	defaultApp = "eureka"

	// The name is limited to "Netflix" or "Amazon" or "MyOwn"
	// by the enumeration type in the interface of eureka
	// "com.netflix.appinfo.DataCenterInfo".
	// Here, "MyOwn" is used as the default.
	defaultDataCenterInfoName = "MyOwn"

	// Class is used the static variable "MY_DATA_CENTER_INFO_TYPE_MARKER"
	// defined in eureka "com.netflix.discovery.converters.jackson"，
	// that is kept for backward compatibility
	defaultDataCenterInfoClass = "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo"

	expansionDatasource = "datasource"
)

// toSyncData transform eureka service cache to SyncData
func toSyncData(eureka *eurekaData) (data *pb.SyncData) {
	apps := eureka.APPS.Applications
	data = &pb.SyncData{
		Services:  make([]*pb.SyncService, 0, len(apps)),
		Instances: make([]*pb.SyncInstance, 0, 10),
	}

	for _, app := range apps {
		service := toSyncService(app)

		instances := toSyncInstances(service.ServiceId, app.Instances)
		if len(instances) == 0 {
			continue
		}

		data.Services = append(data.Services, service)
		data.Instances = append(data.Instances, instances...)
	}
	return
}

// toSyncService transform eureka application to SyncService
func toSyncService(app *Application) (service *pb.SyncService) {
	appName := strings.ToLower(app.Name)
	service = &pb.SyncService{
		ServiceId:     appName,
		Name:          appName,
		App:           defaultApp,
		Version:       "0.0.1",
		DomainProject: "default/default",
		Status:        pb.SyncService_UP,
		PluginName:    PluginName,
	}
	return
}

//  toSyncInstances transform eureka instances to SyncInstances
func toSyncInstances(serviceID string, instances []*Instance) []*pb.SyncInstance {
	instList := make([]*pb.SyncInstance, 0, len(instances))
	for _, inst := range instances {
		if inst.Status != UP {
			continue
		}

		instList = append(instList, toSyncInstance(serviceID, inst))
	}
	return instList
}

// toSyncInstance transform eureka instance to SyncInstance
func toSyncInstance(serviceID string, instance *Instance) (syncInstance *pb.SyncInstance) {
	syncInstance = &pb.SyncInstance{
		InstanceId: instance.InstanceID,
		ServiceId:  serviceID,
		Endpoints:  make([]string, 0, 2),
		HostName:   instance.HostName,
		PluginName: PluginName,
		Version:    "latest",
	}

	switch instance.Status {
	case UP:
		syncInstance.Status = pb.SyncInstance_UP
	case DOWN:
		syncInstance.Status = pb.SyncInstance_DOWN
	case STARTING:
		syncInstance.Status = pb.SyncInstance_STARTING
	case OUTOFSERVICE:
		syncInstance.Status = pb.SyncInstance_OUTOFSERVICE
	default:
		syncInstance.Status = pb.SyncInstance_UNKNOWN
	}

	if instance.Port.Enabled.Bool() {
		syncInstance.Endpoints = append(syncInstance.Endpoints, fmt.Sprintf("http://%s:%d", instance.IPAddr, instance.Port.Port))
	}

	if instance.SecurePort.Enabled.Bool() {
		syncInstance.Endpoints = append(syncInstance.Endpoints, fmt.Sprintf("https://%s:%d", instance.IPAddr, instance.SecurePort.Port))
	}

	if instance.HealthCheckURL != "" {
		syncInstance.HealthCheck = &pb.HealthCheck{
			Interval: 30,
			Times:    3,
			Url:      instance.HealthCheckURL,
		}
	}

	content, err := json.Marshal(instance)
	if err != nil {
		log.Errorf(err, "transform sc service to syncer service failed: %s", err)
		return
	}
	syncInstance.Expansions = []*pb.Expansion{{
		Kind:   expansionDatasource,
		Bytes:  content,
		Labels: map[string]string{},
	}}
	return
}

// toInstance transform SyncInstance to eureka instance
func toInstance(serviceID string, syncInstance *pb.SyncInstance) (instance *Instance) {
	instance = &Instance{}
	if syncInstance.PluginName == PluginName && len(syncInstance.Expansions) > 0 {
		matches := pb.Expansions(syncInstance.Expansions).Find(expansionDatasource, map[string]string{})
		if len(matches) > 0 {
			err := json.Unmarshal(matches[0].Bytes, instance)
			if err == nil {
				return
			}
			log.Errorf(err, "proto unmarshal %s instance, instanceID = %s, kind = %v, content = %v failed",
				PluginName, syncInstance.InstanceId, matches[0].Kind, matches[0].Bytes)
		}
	}
	instance.InstanceID = syncInstance.InstanceId
	instance.APP = serviceID
	instance.Status = pb.SyncInstance_Status_name[int32(syncInstance.Status)]
	instance.OverriddenStatus = UNKNOWN
	instance.DataCenterInfo = &DataCenterInfo{
		Name:  defaultDataCenterInfoName,
		Class: defaultDataCenterInfoClass,
	}

	instance.Metadata = &MetaData{
		Map: map[string]string{"instanceId": syncInstance.InstanceId},
	}

	ipAddr := ""
	for _, ep := range syncInstance.Endpoints {
		addr, err := url.Parse(ep)
		if err != nil {
			log.Error("parse the endpoint of eureka`s instance failed", err)
			continue
		}
		hostname := addr.Hostname()
		if ipAddr != "" || ipAddr != hostname {
			log.Error("eureka`s ipAddr must be unique in endpoints", err)
		}
		ipAddr = hostname

		port := &Port{}
		port.Port, err = strconv.Atoi(addr.Port())
		if err != nil {
			log.Error("Illegal value of port", err)
			continue
		}
		port.Enabled.Set(true)

		switch addr.Scheme {
		case "http":
			instance.Port = port
			instance.VipAddress = serviceID
		case "https":
			instance.SecurePort = port
			instance.SecureVipAddress = ep
		}
	}
	log.Error(ipAddr, errors.New("sc to eureka hostname failed"))
	instance.IPAddr = ipAddr
	instance.HostName = ipAddr

	if syncInstance.HealthCheck != nil {
		instance.HealthCheckURL = syncInstance.HealthCheck.Url
	}

	instArr := strings.Split(syncInstance.InstanceId, ":")
	if len(instArr) != 3 {
		if instance.Port != nil && instance.Port.Enabled.Bool() {
			instance.InstanceID = fmt.Sprintf(eurekaInstanceFormat, ipAddr, instance.APP, instance.Port.Port)
		} else if instance.SecurePort != nil && instance.SecurePort.Enabled.Bool() {
			instance.InstanceID = fmt.Sprintf(eurekaInstanceFormat, ipAddr, instance.APP, instance.SecurePort.Port)
		}
	}
	return
}
