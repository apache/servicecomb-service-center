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
	"regexp"
)

var (
	VersionRegex                  *regexp.Regexp
	MicroServiceValidator         validate.Validator
	MicroServiceInstanceValidator validate.Validator
	ServiceRuleValidator          validate.Validator
	ServicePathValidator          validate.Validator
	HealthCheckInfoValidator      validate.Validator
	MicroServiceKeyValidator      validate.Validator
	DataCenterInfoValidator       validate.Validator
	GetMSExistsReqValidator       validate.Validator
	GetSchemaExistsReqValidator   validate.Validator
	GetServiceReqValidator        validate.Validator
	GetSchemaReqValidator         validate.Validator
	DependencyMSValidator         validate.Validator
	ProviderMsValidator           validate.Validator
	MSDependencyValidator         validate.Validator
	TagReqValidator               validate.Validator
	FindInstanceReqValidator      validate.Validator
	GetInstanceValidator          validate.Validator
)

func init() {
	// 非map/slice的validator
	nameRegex, _ := regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`)
	serviceNameForFindRegex, _ := regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.:]*[a-zA-Z0-9]$`)
	//name模糊规则: name, *
	nameFuzzyRegex, _ := regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$|^\*$`)
	VersionRegex, _ = regexp.Compile(`^[0-9]+(\.[0-9]+)*$`)
	// version模糊规则: 1.0, 1.0+, 1.0-2.0, latest
	versionFuzzyRegex, _ := regexp.Compile(`^[0-9]*$|^[0-9]+(\.[0-9]+)*\+{0,1}$|^[0-9]+(\.[0-9]+)*-[0-9]+(\.[0-9]+)*$|^latest$`)
	pathRegex, _ := regexp.Compile(`^[A-Za-z0-9\.\,\?\'\\/\+&amp;%\$#\=~_\-@\{}]*$`)
	descriptionRegex, _ := regexp.Compile(`^[\p{Han}\w\s。.:\*,\-：”“]*$`)
	levelRegex, _ := regexp.Compile(`^(FRONT|MIDDLE|BACK)*$`)
	statusRegex, _ := regexp.Compile(`^(UP|DOWN)*$`)
	serviceIdRegex, _ := regexp.Compile(`^.*$`)
	aliasRegex, _ := regexp.Compile(`^[a-zA-Z0-9_\-.:]*$`)
	stageRegex, _ := regexp.Compile("^(" + beego.AppConfig.String("stage_rules") + ")*$")
	// map/slice元素的validator
	// 元素的格式和长度由正则控制
	// map/slice的长度由validator中的min/max/length控制
	schemaIdRegex, _ := regexp.Compile(`^[a-zA-Z0-9]{1,160}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{0,158}[a-zA-Z0-9]$`) //length:{1,160}
	instStatusRegex, _ := regexp.Compile(`^(UP|DOWN|STARTING|OUTOFSERVICE)$`)
	tagRegex, _ := regexp.Compile(`^[a-zA-Z][a-zA-Z0-9_\-\.]{0,63}$`)
	hbModeRegex, _ := regexp.Compile(`^(push|pull)$`)
	numberAllowEmptyRegex, _ := regexp.Compile(`^[0-9]*$`)
	numberRegex, _ := regexp.Compile(`^[0-9]+$`)
	epRegex, _ := regexp.Compile(`^[A-Za-z0-9:/?=&%_.-]+$`)
	simpleNameAllowEmptyRegex, _ := regexp.Compile(`^[A-Za-z0-9_.-]*$`)
	simpleNameRegex, _ := regexp.Compile(`^[A-Za-z0-9_.-]+$`)
	regionRegex, _ := regexp.Compile(`([A-Za-z0-9]+-)+([A-Za-z0-9]+)$`)
	ruleRegex, _ := regexp.Compile(`^(WHITE|BLACK)$`)
	ruleAttrRegex, _ := regexp.Compile(`(^tag_(.*)|(^ServiceId$)|(^AppId$)|(^ServiceName$)|(^Version$)|(^Description$)|(^Level$)|(^Status$))`)

	ServiceIdRule := &validate.ValidateRule{Min: 1, Length: 64, Regexp: serviceIdRegex}
	InstanceStatusRule := &validate.ValidateRule{Regexp: instStatusRegex}
	SchemaIdRule := &validate.ValidateRule{Regexp: schemaIdRegex}
	stageRule := &validate.ValidateRule{Regexp: stageRegex}
	nameRule := &validate.ValidateRule{Min: 1, Max: 128, Regexp: nameRegex}
	versionFuzzyRule := &validate.ValidateRule{Min: 1, Max: 128, Regexp: versionFuzzyRegex}
	tagRule := &validate.ValidateRule{Regexp: tagRegex}

	MicroServiceKeyValidator.AddRule("AppId", &validate.ValidateRule{Min: 1, Max: 160, Regexp: nameRegex})
	MicroServiceKeyValidator.AddRule("ServiceName", nameRule)
	MicroServiceKeyValidator.AddRule("Version", &validate.ValidateRule{Min: 1, Max: 64, Regexp: VersionRegex})

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
	ProviderMsValidator.AddRule("ServiceName", &validate.ValidateRule{Min: 1, Max: 128, Regexp: nameFuzzyRegex})
	ProviderMsValidator.AddRule("Version", versionFuzzyRule)
	ProviderMsValidator.AddRule("Stage", stageRule)

	MSDependencyValidator.AddSub("Consumer", &DependencyMSValidator)
	MSDependencyValidator.AddSub("Providers", &ProviderMsValidator)

	TagReqValidator.AddRule("ServiceId", ServiceIdRule)
	TagReqValidator.AddRule("Tags", tagRule)

	HealthCheckInfoValidator.AddRule("Mode", &validate.ValidateRule{Regexp: hbModeRegex})
	HealthCheckInfoValidator.AddRule("Port", &validate.ValidateRule{Max: math.MaxInt16, Regexp: numberAllowEmptyRegex})
	HealthCheckInfoValidator.AddRule("Times", &validate.ValidateRule{Max: math.MaxInt32, Regexp: numberRegex})
	HealthCheckInfoValidator.AddRule("Interval", &validate.ValidateRule{Max: math.MaxInt32, Regexp: numberRegex})
	HealthCheckInfoValidator.AddRule("Url", &validate.ValidateRule{Regexp: pathRegex})

	MicroServiceInstanceValidator.AddRule("InstanceId", &validate.ValidateRule{Length: 64, Regexp: simpleNameAllowEmptyRegex})
	MicroServiceInstanceValidator.AddRule("ServiceId", ServiceIdRule)
	MicroServiceInstanceValidator.AddRule("Endpoints", &validate.ValidateRule{Regexp: epRegex})
	MicroServiceInstanceValidator.AddRule("HostName", &validate.ValidateRule{Length: 64, Regexp: simpleNameRegex})
	MicroServiceInstanceValidator.AddSub("HealthCheck", &HealthCheckInfoValidator)
	MicroServiceInstanceValidator.AddRule("Status", InstanceStatusRule)
	MicroServiceInstanceValidator.AddRule("Stage", &validate.ValidateRule{Min: 1, Regexp: stageRegex})
	MicroServiceInstanceValidator.AddSub("DataCenterInfo", &DataCenterInfoValidator)

	DataCenterInfoValidator.AddRule("Name", &validate.ValidateRule{Length: 128, Regexp: simpleNameRegex})
	DataCenterInfoValidator.AddRule("Region", &validate.ValidateRule{Length: 128, Regexp: regionRegex})
	DataCenterInfoValidator.AddRule("AvailableZone", &validate.ValidateRule{Length: 128, Regexp: regionRegex})

	ServiceRuleValidator.AddRule("RuleType", &validate.ValidateRule{Regexp: ruleRegex})
	ServiceRuleValidator.AddRule("Attribute", &validate.ValidateRule{Regexp: ruleAttrRegex})
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
	GetInstanceValidator.AddRule("ProviderInstanceId", &validate.ValidateRule{Min: 1, Max: 64, Regexp: simpleNameAllowEmptyRegex})
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
	case *pb.MicroServiceInstance, *pb.UpdateInstanceStatusRequest:
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
