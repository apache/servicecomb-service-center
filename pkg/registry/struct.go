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

package registry

import "github.com/apache/servicecomb-service-center/pkg/types"

const (
	MS_UP   string = "UP"
	MS_DOWN string = "DOWN"

	MSI_UP           string = "UP"
	MSI_DOWN         string = "DOWN"
	MSI_STARTING     string = "STARTING"
	MSI_TESTING      string = "TESTING"
	MSI_OUTOFSERVICE string = "OUTOFSERVICE"

	ENV_DEV    string = "development"
	ENV_TEST   string = "testing"
	ENV_ACCEPT string = "acceptance"
	ENV_PROD   string = "production"

	REGISTERBY_SDK      string = "SDK"
	REGISTERBY_SIDECAR  string = "SIDECAR"
	REGISTERBY_PLATFORM string = "PLATFORM"

	EVT_INIT   types.EventType = "INIT"
	EVT_CREATE types.EventType = "CREATE"
	EVT_UPDATE types.EventType = "UPDATE"
	EVT_DELETE types.EventType = "DELETE"
	EVT_EXPIRE types.EventType = "EXPIRE"
	EVT_ERROR  types.EventType = "ERROR"

	CHECK_BY_HEARTBEAT string = "push"
	CHECK_BY_PLATFORM  string = "pull"
)

type HeartbeatSetRequest struct {
	Instances []*HeartbeatSetElement `protobuf:"bytes,1,rep,name=instances" json:"instances,omitempty"`
}

type HeartbeatSetElement struct {
	ServiceId  string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instanceId" json:"instanceId,omitempty"`
}

type HeartbeatSetResponse struct {
	Response  *Response        `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Instances []*InstanceHbRst `protobuf:"bytes,2,rep,name=instances" json:"instances,omitempty"`
}

type InstanceHbRst struct {
	ServiceId  string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instanceId" json:"instanceId,omitempty"`
	ErrMessage string `protobuf:"bytes,3,opt,name=errMessage" json:"errMessage,omitempty"`
}

type StService struct {
	Count       int64 `protobuf:"varint,1,opt,name=count" json:"count,omitempty"`
	OnlineCount int64 `protobuf:"varint,2,opt,name=onlineCount" json:"onlineCount,omitempty"`
}

type StInstance struct {
	Count         int64 `protobuf:"varint,1,opt,name=count" json:"count,omitempty"`
	CountByDomain int64 `protobuf:"varint,2,opt,name=countByDomain" json:"countByDomain,omitempty"`
}

type StApp struct {
	Count int64 `protobuf:"varint,1,opt,name=count" json:"count,omitempty"`
}

type Statistics struct {
	Services  *StService  `protobuf:"bytes,1,opt,name=services" json:"services,omitempty"`
	Instances *StInstance `protobuf:"bytes,2,opt,name=instances" json:"instances,omitempty"`
	Apps      *StApp      `protobuf:"bytes,3,opt,name=apps" json:"apps,omitempty"`
}

type GetServicesInfoRequest struct {
	Options     []string `protobuf:"bytes,1,rep,name=options" json:"options,omitempty"`
	AppId       string   `protobuf:"bytes,2,opt,name=appId" json:"appId,omitempty"`
	ServiceName string   `protobuf:"bytes,3,opt,name=serviceName" json:"serviceName,omitempty"`
	CountOnly   bool     `protobuf:"varint,4,opt,name=countOnly" json:"countOnly,omitempty"`
	WithShared  bool     `protobuf:"varint,5,opt,name=withShared" json:"withShared,omitempty"`
}

type GetServicesInfoResponse struct {
	Response          *Response        `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	AllServicesDetail []*ServiceDetail `protobuf:"bytes,2,rep,name=allServicesDetail" json:"allServicesDetail,omitempty"`
	Statistics        *Statistics      `protobuf:"bytes,3,opt,name=statistics" json:"statistics,omitempty"`
}

type MicroServiceKey struct {
	// Tenant: The format is "{domain}/{project}"
	Tenant string `protobuf:"bytes,1,opt,name=tenant" json:"tenant,omitempty"`
	// Deprecated: Use Tenant instead
	Project     string `protobuf:"bytes,2,opt,name=project" json:"project,omitempty"`
	AppId       string `protobuf:"bytes,3,opt,name=appId" json:"appId,omitempty"`
	ServiceName string `protobuf:"bytes,4,opt,name=serviceName" json:"serviceName,omitempty"`
	Version     string `protobuf:"bytes,5,opt,name=version" json:"version,omitempty"`
	Environment string `protobuf:"bytes,6,opt,name=environment" json:"environment,omitempty"`
	Alias       string `protobuf:"bytes,7,opt,name=alias" json:"alias,omitempty"`
}

type FrameWorkProperty struct {
	Name    string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Version string `protobuf:"bytes,2,opt,name=version" json:"version,omitempty"`
}

type ServiceRule struct {
	RuleId       string `protobuf:"bytes,1,opt,name=ruleId" json:"ruleId,omitempty"`
	RuleType     string `protobuf:"bytes,2,opt,name=ruleType" json:"ruleType,omitempty"`
	Attribute    string `protobuf:"bytes,3,opt,name=attribute" json:"attribute,omitempty"`
	Pattern      string `protobuf:"bytes,4,opt,name=pattern" json:"pattern,omitempty"`
	Description  string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	Timestamp    string `protobuf:"bytes,6,opt,name=timestamp" json:"timestamp,omitempty"`
	ModTimestamp string `protobuf:"bytes,7,opt,name=modTimestamp" json:"modTimestamp,omitempty"`
}

type AddOrUpdateServiceRule struct {
	RuleType    string `protobuf:"bytes,1,opt,name=ruleType" json:"ruleType,omitempty"`
	Attribute   string `protobuf:"bytes,2,opt,name=attribute" json:"attribute,omitempty"`
	Pattern     string `protobuf:"bytes,3,opt,name=pattern" json:"pattern,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
}

type ServicePath struct {
	Path     string            `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Property map[string]string `protobuf:"bytes,2,rep,name=property" json:"property,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

type Response struct {
	Code    int32  `protobuf:"varint,1,opt,name=code" json:"code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *Response) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}
func (m *Response) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type GetExistenceByIDRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
}

type GetExistenceRequest struct {
	Type        string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	AppId       string `protobuf:"bytes,2,opt,name=appId" json:"appId,omitempty"`
	ServiceName string `protobuf:"bytes,3,opt,name=serviceName" json:"serviceName,omitempty"`
	Version     string `protobuf:"bytes,4,opt,name=version" json:"version,omitempty"`
	ServiceId   string `protobuf:"bytes,5,opt,name=serviceId" json:"serviceId,omitempty"`
	SchemaId    string `protobuf:"bytes,6,opt,name=schemaId" json:"schemaId,omitempty"`
	Environment string `protobuf:"bytes,7,opt,name=environment" json:"environment,omitempty"`
}

type GetExistenceByIDResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Exist    bool      `protobuf:"bytes,2,opt,name=exist" json:"exist,omitempty"`
}

type GetExistenceResponse struct {
	Response  *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	ServiceId string    `protobuf:"bytes,2,opt,name=serviceId" json:"serviceId,omitempty"`
	SchemaId  string    `protobuf:"bytes,3,opt,name=schemaId" json:"schemaId,omitempty"`
	Summary   string    `protobuf:"bytes,4,opt,name=summary" json:"summary,omitempty"`
}

func (m *GetExistenceResponse) GetSummary() string {
	if m != nil {
		return m.Summary
	}
	return ""
}

type CreateServiceRequest struct {
	Service   *MicroService             `protobuf:"bytes,1,opt,name=service" json:"service,omitempty"`
	Rules     []*AddOrUpdateServiceRule `protobuf:"bytes,2,rep,name=rules" json:"rules,omitempty"`
	Tags      map[string]string         `protobuf:"bytes,3,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Instances []*MicroServiceInstance   `protobuf:"bytes,4,rep,name=instances" json:"instances,omitempty"`
}

type CreateServiceResponse struct {
	Response  *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	ServiceId string    `protobuf:"bytes,2,opt,name=serviceId" json:"serviceId,omitempty"`
}

type DeleteServiceRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Force     bool   `protobuf:"varint,2,opt,name=force" json:"force,omitempty"`
}

type DeleteServiceResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type GetServiceRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
}

type GetServiceResponse struct {
	Response *Response     `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Service  *MicroService `protobuf:"bytes,2,opt,name=service" json:"service,omitempty"`
}

type GetServicesRequest struct {
}

type GetServicesResponse struct {
	Response *Response       `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Services []*MicroService `protobuf:"bytes,2,rep,name=services" json:"services,omitempty"`
}

type UpdateServicePropsRequest struct {
	ServiceId  string            `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Properties map[string]string `protobuf:"bytes,2,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

type UpdateServicePropsResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type GetServiceRulesRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
}

type GetServiceRulesResponse struct {
	Response *Response      `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Rules    []*ServiceRule `protobuf:"bytes,2,rep,name=rules" json:"rules,omitempty"`
}

type UpdateServiceRuleRequest struct {
	ServiceId string                  `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	RuleId    string                  `protobuf:"bytes,2,opt,name=ruleId" json:"ruleId,omitempty"`
	Rule      *AddOrUpdateServiceRule `protobuf:"bytes,3,opt,name=rule" json:"rule,omitempty"`
}

type UpdateServiceRuleResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type AddServiceRulesRequest struct {
	ServiceId string                    `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Rules     []*AddOrUpdateServiceRule `protobuf:"bytes,2,rep,name=rules" json:"rules,omitempty"`
}

type AddServiceRulesResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	RuleIds  []string  `protobuf:"bytes,2,rep,name=RuleIds" json:"RuleIds,omitempty"`
}

type DeleteServiceRulesRequest struct {
	ServiceId string   `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	RuleIds   []string `protobuf:"bytes,2,rep,name=ruleIds" json:"ruleIds,omitempty"`
}

type DeleteServiceRulesResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type GetServiceTagsRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
}

type GetServiceTagsResponse struct {
	Response *Response         `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Tags     map[string]string `protobuf:"bytes,2,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

type UpdateServiceTagRequest struct {
	ServiceId string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	Value     string `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
}

type UpdateServiceTagResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type AddServiceTagsRequest struct {
	ServiceId string            `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Tags      map[string]string `protobuf:"bytes,2,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

type AddServiceTagsResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type DeleteServiceTagsRequest struct {
	ServiceId string   `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	Keys      []string `protobuf:"bytes,2,rep,name=keys" json:"keys,omitempty"`
}

type DeleteServiceTagsResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type HealthCheck struct {
	Mode     string `protobuf:"bytes,1,opt,name=mode" json:"mode,omitempty"`
	Port     int32  `protobuf:"varint,2,opt,name=port" json:"port,omitempty"`
	Interval int32  `protobuf:"varint,3,opt,name=interval" json:"interval,omitempty"`
	Times    int32  `protobuf:"varint,4,opt,name=times" json:"times,omitempty"`
	Url      string `protobuf:"bytes,5,opt,name=url" json:"url,omitempty"`
}

type DataCenterInfo struct {
	Name          string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Region        string `protobuf:"bytes,2,opt,name=region" json:"region,omitempty"`
	AvailableZone string `protobuf:"bytes,3,opt,name=availableZone" json:"availableZone,omitempty"`
}

type MicroServiceInstanceKey struct {
	InstanceId string `protobuf:"bytes,1,opt,name=instanceId" json:"instanceId,omitempty"`
	ServiceId  string `protobuf:"bytes,2,opt,name=serviceId" json:"serviceId,omitempty"`
}

type WatchInstanceRequest struct {
	SelfServiceId string `protobuf:"bytes,1,opt,name=selfServiceId" json:"selfServiceId,omitempty"`
}

type WatchInstanceResponse struct {
	Response *Response             `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Action   string                `protobuf:"bytes,2,opt,name=action" json:"action,omitempty"`
	Key      *MicroServiceKey      `protobuf:"bytes,3,opt,name=key" json:"key,omitempty"`
	Instance *MicroServiceInstance `protobuf:"bytes,4,opt,name=instance" json:"instance,omitempty"`
}

type AddDependenciesRequest struct {
	Dependencies []*ConsumerDependency `protobuf:"bytes,1,rep,name=dependencies" json:"dependencies,omitempty"`
}

type AddDependenciesResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type CreateDependenciesRequest struct {
	Dependencies []*ConsumerDependency `protobuf:"bytes,1,rep,name=dependencies" json:"dependencies,omitempty"`
}

type ConsumerDependency struct {
	Consumer  *MicroServiceKey   `protobuf:"bytes,1,opt,name=consumer" json:"consumer,omitempty"`
	Providers []*MicroServiceKey `protobuf:"bytes,2,rep,name=providers" json:"providers,omitempty"`
	Override  bool               `protobuf:"varint,3,opt,name=override" json:"override,omitempty"`
}

type CreateDependenciesResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

type GetDependenciesRequest struct {
	ServiceId  string `protobuf:"bytes,1,opt,name=serviceId" json:"serviceId,omitempty"`
	SameDomain bool   `protobuf:"varint,2,opt,name=sameDomain" json:"sameDomain,omitempty"`
	NoSelf     bool   `protobuf:"varint,3,opt,name=noSelf" json:"noSelf,omitempty"`
}

type GetConDependenciesResponse struct {
	Response  *Response       `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Providers []*MicroService `protobuf:"bytes,2,rep,name=providers" json:"providers,omitempty"`
}

type GetProDependenciesResponse struct {
	Response  *Response       `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Consumers []*MicroService `protobuf:"bytes,2,rep,name=consumers" json:"consumers,omitempty"`
}

// 服务详情
type ServiceDetail struct {
	MicroService         *MicroService           `protobuf:"bytes,1,opt,name=microService" json:"microService,omitempty"`
	Instances            []*MicroServiceInstance `protobuf:"bytes,2,rep,name=instances" json:"instances,omitempty"`
	SchemaInfos          []*Schema               `protobuf:"bytes,3,rep,name=schemaInfos" json:"schemaInfos,omitempty"`
	Rules                []*ServiceRule          `protobuf:"bytes,4,rep,name=rules" json:"rules,omitempty"`
	Providers            []*MicroService         `protobuf:"bytes,5,rep,name=providers" json:"providers,omitempty"`
	Consumers            []*MicroService         `protobuf:"bytes,6,rep,name=consumers" json:"consumers,omitempty"`
	Tags                 map[string]string       `protobuf:"bytes,7,rep,name=tags" json:"tags,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	MicroServiceVersions []string                `protobuf:"bytes,8,rep,name=microServiceVersions" json:"microServiceVersions,omitempty"`
	Statics              *Statistics             `protobuf:"bytes,9,opt,name=statics" json:"statics,omitempty"`
}

// 服务详情返回信息
type GetServiceDetailResponse struct {
	Response *Response      `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Service  *ServiceDetail `protobuf:"bytes,2,opt,name=service" json:"service,omitempty"`
}

// 删除服务响应内容
type DelServicesRspInfo struct {
	ErrMessage string `protobuf:"bytes,1,opt,name=errMessage" json:"errMessage,omitempty"`
	ServiceId  string `protobuf:"bytes,2,opt,name=serviceId" json:"serviceId,omitempty"`
}

// 删除服务响应
type DelServicesResponse struct {
	Response *Response             `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Services []*DelServicesRspInfo `protobuf:"bytes,2,rep,name=services" json:"services,omitempty"`
}

type GetAppsRequest struct {
	Environment string `protobuf:"bytes,1,opt,name=environment" json:"environment,omitempty"`
	WithShared  bool   `protobuf:"varint,2,opt,name=withShared" json:"withShared,omitempty"`
}

type GetAppsResponse struct {
	Response *Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	AppIds   []string  `protobuf:"bytes,2,rep,name=appIds" json:"appIds,omitempty"`
}
type MicroServiceDependency struct {
	Dependency []*MicroServiceKey `json:"Dependency,omitempty"`
}

type BatchGetInstancesRequest struct {
	ServiceIds []string `json:"serviceIds,omitempty"`
}
