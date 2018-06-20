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
package service

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/pkg/validate"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"regexp"
)

var (
	microServiceKeyValidator       validate.Validator
	existenceReqValidator          validate.Validator
	getServiceReqValidator         validate.Validator
	createServiceReqValidator      validate.Validator
	updateServicePropsReqValidator validate.Validator
)

var (
	// 非map/slice的validator
	nameRegex, _ = regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`)
	// find 支持alias，多个:
	serviceNameForFindRegex, _ = regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.:]*[a-zA-Z0-9]$`)
	// version規則: x[.y[.z]]
	versionRegex = serviceUtil.NewVersionRegexp(false)
	// version模糊规则: 1.0, 1.0+, 1.0-2.0, latest
	versionFuzzyRegex  = serviceUtil.NewVersionRegexp(true)
	pathRegex, _       = regexp.Compile(`^[A-Za-z0-9.,?'\\/+&amp;%$#=~_\-@{}]*$`)
	levelRegex, _      = regexp.Compile(`^(FRONT|MIDDLE|BACK)$`)
	statusRegex, _     = regexp.Compile("^(" + pb.MS_UP + "|" + pb.MS_DOWN + ")?$")
	serviceIdRegex, _  = regexp.Compile(`^\S*$`)
	aliasRegex, _      = regexp.Compile(`^[a-zA-Z0-9_\-.:]*$`)
	registerByRegex, _ = regexp.Compile("^(" + util.StringJoin([]string{pb.REGISTERBY_SDK, pb.REGISTERBY_SIDECAR}, "|") + ")*$")
	envRegex, _        = regexp.Compile("^(" + util.StringJoin([]string{
		pb.ENV_DEV, pb.ENV_TEST, pb.ENV_ACCEPT, pb.ENV_PROD}, "|") + ")*$")
	schemaIdRegex, _ = regexp.Compile(`^[a-zA-Z0-9]{1,160}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{0,158}[a-zA-Z0-9]$`)
)

func MicroServiceKeyValidator() *validate.Validator {
	return microServiceKeyValidator.Init(func(v *validate.Validator) {
		v.AddRule("Environment", &validate.ValidateRule{Regexp: envRegex})
		v.AddRule("AppId", &validate.ValidateRule{Min: 1, Max: 160, Regexp: nameRegex})
		v.AddRule("ServiceName", &validate.ValidateRule{Min: 1, Max: 128, Regexp: nameRegex})
		v.AddRule("Version", &validate.ValidateRule{Min: 1, Max: 64, Regexp: versionRegex})
	})
}

func ExistenceReqValidator() *validate.Validator {
	return existenceReqValidator.Init(func(v *validate.Validator) {
		v.AddRules(MicroServiceKeyValidator().GetRules())
		v.AddRule("ServiceName", &validate.ValidateRule{Min: 1, Max: 160 + 1 + 128, Regexp: serviceNameForFindRegex})
		v.AddRule("Version", &validate.ValidateRule{Min: 1, Max: 129, Regexp: versionFuzzyRegex})
	})
}

func GetServiceReqValidator() *validate.Validator {
	return getServiceReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", &validate.ValidateRule{Min: 1, Max: 64, Regexp: serviceIdRegex})
	})
}

func CreateServiceReqValidator() *validate.Validator {
	return createServiceReqValidator.Init(func(v *validate.Validator) {
		var pathValidator validate.Validator
		pathValidator.AddRule("Path", &validate.ValidateRule{Regexp: pathRegex})

		var frameworkValidator validate.Validator
		frameworkValidator.AddRule("Name", &validate.ValidateRule{Max: 64, Regexp: nameRegex})
		frameworkValidator.AddRule("Version", &validate.ValidateRule{Max: 64})

		var microServiceValidator validate.Validator
		microServiceValidator.AddRules(MicroServiceKeyValidator().GetRules())
		microServiceValidator.AddRule("AppId", &validate.ValidateRule{Max: 160, Regexp: nameRegex})
		microServiceValidator.AddRule("Version", &validate.ValidateRule{Max: 64, Regexp: versionRegex})
		microServiceValidator.AddRule("ServiceId", &validate.ValidateRule{Max: 64, Regexp: serviceIdRegex})
		microServiceValidator.AddRule("Description", &validate.ValidateRule{Max: 256})
		microServiceValidator.AddRule("Level", &validate.ValidateRule{Regexp: levelRegex})
		microServiceValidator.AddRule("Status", &validate.ValidateRule{Regexp: statusRegex})
		microServiceValidator.AddRule("Schemas", &validate.ValidateRule{Max: quota.DefaultSchemaQuota, Regexp: schemaIdRegex})
		microServiceValidator.AddSub("Paths", &pathValidator)
		microServiceValidator.AddRule("Alias", &validate.ValidateRule{Max: 128, Regexp: aliasRegex})
		microServiceValidator.AddRule("RegisterBy", &validate.ValidateRule{Max: 64, Regexp: registerByRegex})
		microServiceValidator.AddSub("Framework", &frameworkValidator)

		v.AddRule("Service", &validate.ValidateRule{Min: 1})
		v.AddSub("Service", &microServiceValidator)
	})

}

func UpdateServicePropsReqValidator() *validate.Validator {
	return updateServicePropsReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Properties", &validate.ValidateRule{Min: 1})
	})
}
