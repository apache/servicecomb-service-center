//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package core

import (
	"errors"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/validate"
	"github.com/astaxie/beego"
	"math"
	"reflect"
)

var (
	MicroServiceValidator         validate.Validator
	MicroServiceInstanceValidator validate.Validator
	ServiceRuleValidator          validate.Validator
	ServicePathValidator          validate.Validator
	ServiceIdRule                 *validate.ValidateRule
	InstanseStatusRule            *validate.ValidateRule
	HealthCheckInfoValidator      validate.Validator
	MicroServiceKeyValidator      validate.Validator
	DataCenterInfoValidator       validate.Validator
	GetMSExistsReqValidator       validate.Validator
	GetSchemaExistsReqValidator   validate.Validator
	GetServiceReqValidator        validate.Validator
	GetSchemaReqValidator         validate.Validator
	SchemaIdRule                  *validate.ValidateRule
	DependencyMSValidator         validate.Validator
	ProviderMsValidator           validate.Validator
	MSDependencyValidator         validate.Validator
	TagReqValidator               validate.Validator
	FindInstanceReqValidator      validate.Validator
	GetInstanceValidator          validate.Validator
)

func init() {
	// 非map/slice的validator
	nameRegex := `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`
	serviceNameForFindRegex := `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.:]*[a-zA-Z0-9]$`
	//name模糊规则: name, *
	nameFuzzyRegex := `^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$|^\*$`
	versionRegex := `^[0-9]*$|^[0-9]+(\.[0-9]+)*$`
	// version模糊规则: 1.0, 1.0+, 1.0-2.0, latest
	versionFuzzyRegex := `^[0-9]*$|^[0-9]+(\.[0-9]+)*\+{0,1}$|^[0-9]+(\.[0-9]+)*-[0-9]+(\.[0-9]+)*$|^latest$`
	pathRegex := `^[A-Za-z0-9\.\,\?\'\\/\+&amp;%\$#\=~_\-@\{}]*$`
	descriptionRegex := `^[\p{Han}\w\s。.:\*,\-：”“]*$`
	levelRegex := `^(FRONT|MIDDLE|BACK)*$`
	statusRegex := `^(UP|DOWN)*$`
	serviceIdRegex := `^.*$`
	aliasRegex := `^[a-zA-Z0-9_\-.:]*$`
	stageRegex := "^(" + beego.AppConfig.String("stage_rules") + ")*$"
	instanceIdRegex := `^[A-Za-z0-9_.-]*$`

	// map/slice元素的validator
	// 元素的格式和长度由正则控制
	// map/slice的长度由validator中的min/max/length控制
	schemaIdRegex := `^[a-zA-Z0-9]{1,160}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{0,158}[a-zA-Z0-9]$` //length:{1,160}

	ServiceIdRule = &validate.ValidateRule{Min: 1, Length: 64, Regexp: serviceIdRegex}
	InstanseStatusRule = &validate.ValidateRule{Regexp: `^(UP|DOWN|STARTING|OUTOFSERVICE)$`}
	SchemaIdRule = &validate.ValidateRule{Regexp: schemaIdRegex}
	stageRule := &validate.ValidateRule{Regexp: stageRegex}
	nameRule := &validate.ValidateRule{Min: 1, Max: 128, Regexp: nameRegex}
	nameFuzzyRule := &validate.ValidateRule{Min: 1, Max: 128, Regexp: nameFuzzyRegex}
	versionFuzzyRule := &validate.ValidateRule{Min: 1, Max: 128, Regexp: versionFuzzyRegex}
	tagRule := &validate.ValidateRule{Regexp: `^[a-zA-Z][a-zA-Z0-9_\-\.]{0,63}$`}

	MicroServiceKeyValidator.AddRule("AppId", &validate.ValidateRule{Min: 1, Max: 160, Regexp: nameRegex})
	MicroServiceKeyValidator.AddRule("ServiceName", nameRule)
	MicroServiceKeyValidator.AddRule("Version", &validate.ValidateRule{Min: 1, Max: 64, Regexp: versionRegex})

	ServicePathValidator.AddRule("Path", &validate.ValidateRule{Regexp: pathRegex})

	MicroServiceValidator.AddRules(MicroServiceKeyValidator.GetRules())
	MicroServiceValidator.AddRule("Description", &validate.ValidateRule{Length: 256, Regexp: descriptionRegex})
	MicroServiceValidator.AddRule("Level", &validate.ValidateRule{Min: 1, Regexp: levelRegex})
	MicroServiceValidator.AddRule("Status", &validate.ValidateRule{Min: 1, Regexp: statusRegex})
	MicroServiceValidator.AddRule("Schemas", SchemaIdRule)
	MicroServiceValidator.AddSub("Paths", &ServicePathValidator)
	MicroServiceValidator.AddRule("Alias", &validate.ValidateRule{Length: 128, Regexp: aliasRegex})

	GetMSExistsReqValidator.AddRule("AppId", MicroServiceKeyValidator.GetRule("AppId"))
	GetMSExistsReqValidator.AddRule("ServiceName", nameRule)
	GetMSExistsReqValidator.AddRule("Version", versionFuzzyRule)

	GetSchemaExistsReqValidator.AddRule("ServiceId", ServiceIdRule)
	GetSchemaExistsReqValidator.AddRule("SchemaId", SchemaIdRule)

	GetServiceReqValidator.AddRule("ServiceId", ServiceIdRule)

	GetSchemaReqValidator.AddRule("ServiceId", ServiceIdRule)
	GetSchemaReqValidator.AddRule("SchemaId", SchemaIdRule)

	DependencyMSValidator.AddRules(MicroServiceKeyValidator.GetRules())
	DependencyMSValidator.AddRule("Stage", stageRule)

	ProviderMsValidator.AddRule("AppId", MicroServiceKeyValidator.GetRule("AppId"))
	ProviderMsValidator.AddRule("ServiceName", nameFuzzyRule)
	ProviderMsValidator.AddRule("Version", versionFuzzyRule)
	ProviderMsValidator.AddRule("Stage", stageRule)

	MSDependencyValidator.AddSub("Consumer", &DependencyMSValidator)
	MSDependencyValidator.AddSub("Providers", &ProviderMsValidator)

	TagReqValidator.AddRule("ServiceId", ServiceIdRule)
	TagReqValidator.AddRule("Tags", tagRule)

	HealthCheckInfoValidator.AddRule("Mode", &validate.ValidateRule{Regexp: `^(push|pull)$`})
	HealthCheckInfoValidator.AddRule("Port", &validate.ValidateRule{Max: math.MaxInt16, Regexp: `^[0-9]*$`})
	HealthCheckInfoValidator.AddRule("Times", &validate.ValidateRule{Max: math.MaxInt32, Regexp: `^[0-9]+$`})
	HealthCheckInfoValidator.AddRule("Interval", &validate.ValidateRule{Max: math.MaxInt32, Regexp: `^[0-9]+$`})
	HealthCheckInfoValidator.AddRule("Url", &validate.ValidateRule{Regexp: pathRegex})

	MicroServiceInstanceValidator.AddRule("InstanceId", &validate.ValidateRule{Length: 64, Regexp: instanceIdRegex})
	MicroServiceInstanceValidator.AddRule("ServiceId", ServiceIdRule)
	MicroServiceInstanceValidator.AddRule("Endpoints", &validate.ValidateRule{Regexp: `^[A-Za-z0-9:/?=&%_.-]+$`})
	MicroServiceInstanceValidator.AddRule("HostName", &validate.ValidateRule{Length: 64, Regexp: `^[A-Za-z0-9_.-]+$`})
	MicroServiceInstanceValidator.AddSub("HealthCheck", &HealthCheckInfoValidator)
	MicroServiceInstanceValidator.AddRule("Status", InstanseStatusRule)
	MicroServiceInstanceValidator.AddRule("Stage", &validate.ValidateRule{Min: 1, Regexp: stageRegex})
	MicroServiceInstanceValidator.AddSub("DataCenterInfo", &DataCenterInfoValidator)

	DataCenterInfoValidator.AddRule("Name", &validate.ValidateRule{Length: 128, Regexp: `^[A-Za-z0-9_.-]+$`})
	DataCenterInfoValidator.AddRule("Region", &validate.ValidateRule{Length: 128, Regexp: `([A-Za-z0-9]+-)+([A-Za-z0-9]+)$`})
	DataCenterInfoValidator.AddRule("AvailableZone", &validate.ValidateRule{Length: 128, Regexp: `([A-Za-z0-9]+-)+([A-Za-z0-9]+)$`})

	ServiceRuleValidator.AddRule("RuleType", &validate.ValidateRule{Regexp: `^(WHITE|BLACK)$`})
	ServiceRuleValidator.AddRule("Attribute", &validate.ValidateRule{Regexp: `(^tag_(.*)|(^ServiceId$)|(^AppId$)|(^ServiceName$)|(^Version$)|(^Description$)|(^Level$)|(^Status$))`})
	ServiceRuleValidator.AddRule("Pattern", &validate.ValidateRule{Max: 64, Min: 1})
	ServiceRuleValidator.AddRule("Description", MicroServiceValidator.GetRule("Description"))

	FindInstanceReqValidator.AddRule("ConsumerServiceId", ServiceIdRule)
	FindInstanceReqValidator.AddRule("AppId", MicroServiceKeyValidator.GetRule("AppId"))
	FindInstanceReqValidator.AddRule("ServiceName", &validate.ValidateRule{Min: 1, Max: 128, Regexp: serviceNameForFindRegex})
	FindInstanceReqValidator.AddRule("VersionRule", versionFuzzyRule)
	FindInstanceReqValidator.AddRule("Tags", tagRule)
	FindInstanceReqValidator.AddRule("Stage", stageRule)

	GetInstanceValidator.AddRule("ConsumerServiceId", ServiceIdRule)
	GetInstanceValidator.AddRule("ProviderServiceId", ServiceIdRule)
	GetInstanceValidator.AddRule("ProviderInstanceId", &validate.ValidateRule{Min: 1, Max: 64, Regexp: instanceIdRegex})
	GetInstanceValidator.AddRule("Tags", tagRule)
	GetInstanceValidator.AddRule("Stage", stageRule)
}

func Validate(v interface{}) error {
	if v == nil {
		util.LOGGER.Errorf(nil, "Data is nil!")
		return errors.New("Data is nil!")
	}
	sv := reflect.ValueOf(v)
	if sv.Kind() == reflect.Ptr && sv.IsNil() {
		util.LOGGER.Errorf(nil, "Pointer is nil!")
		return errors.New("Pointer is nil!")
	}
	switch t := v.(type) {
	case (*pb.MicroService):
		return MicroServiceValidator.Validate(v)
	case (*pb.MicroServiceInstance):
		return MicroServiceInstanceValidator.Validate(v)
	case (*pb.AddOrUpdateServiceRule):
		return ServiceRuleValidator.Validate(v)
	case *pb.GetServiceRequest, *pb.UpdateServicePropsRequest,
		*pb.DeleteServiceRequest, *pb.GetDependenciesRequest:
		return GetServiceReqValidator.Validate(v)
	case *pb.AddServiceTagsRequest, *pb.DeleteServiceTagsRequest,
		*pb.UpdateServiceTagRequest, *pb.GetServiceTagsRequest:
		return TagReqValidator.Validate(v)
	case *pb.GetSchemaRequest, *pb.ModifySchemaRequest, *pb.DeleteSchemaRequest:
		return GetSchemaReqValidator.Validate(v)
	case *pb.MicroServiceDependency:
		return DependencyMSValidator.Validate(v)
	case *pb.FindInstancesRequest:
		return FindInstanceReqValidator.Validate(v)
	case *pb.GetOneInstanceRequest, *pb.GetInstancesRequest:
		return GetInstanceValidator.Validate(v)
	default:
		util.LOGGER.Errorf(nil, "No validator for %T.", t)
		return nil
	}
}
