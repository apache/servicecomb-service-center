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

package validator

import (
	"regexp"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/go-chassis/cari/discovery"
)

var (
	microServiceKeyValidator       validate.Validator
	microServiceKeySearchValidator validate.Validator
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
	versionRegex       = validate.NewVersionRegexp(false)
	pathRegex, _       = regexp.Compile(`^[A-Za-z0-9.,?'\\/+&amp;%$#=~_\-@{}]*$`)
	levelRegex, _      = regexp.Compile(`^(FRONT|MIDDLE|BACK)$`)
	statusRegex, _     = regexp.Compile("^(" + discovery.MS_UP + "|" + discovery.MS_DOWN + ")?$")
	serviceIDRegex, _  = regexp.Compile(`^\S*$`)
	aliasRegex, _      = regexp.Compile(`^[a-zA-Z0-9_\-.:]*$`)
	registerByRegex, _ = regexp.Compile("^(" + util.StringJoin([]string{discovery.REGISTERBY_SDK, discovery.REGISTERBY_SIDECAR, discovery.REGISTERBY_PLATFORM}, "|") + ")*$")
	envRegex, _        = regexp.Compile("^(" + util.StringJoin([]string{
		discovery.ENV_DEV, discovery.ENV_TEST, discovery.ENV_ACCEPT, discovery.ENV_PROD}, "|") + ")*$")
	schemaIDRegex, _ = regexp.Compile(`^[a-zA-Z0-9]{1,160}$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]{0,158}[a-zA-Z0-9]$`)

	accountStatusRegex, _ = regexp.Compile(`^(active|inactive)$|^$`)
)

func MicroServiceKeyValidator() *validate.Validator {
	return microServiceKeyValidator.Init(func(v *validate.Validator) {
		v.AddRule("Environment", &validate.Rule{Regexp: envRegex})
		v.AddRule("AppId", &validate.Rule{Min: 1, Max: 160, Regexp: nameRegex})
		v.AddRule("ServiceName", &validate.Rule{Min: 1, Max: 128, Regexp: nameRegex})
		v.AddRule("Version", &validate.Rule{Min: 1, Max: 64, Regexp: versionRegex})
	})
}

func MicroServiceSearchKeyValidator() *validate.Validator {
	return microServiceKeySearchValidator.Init(func(v *validate.Validator) {
		v.AddRule("Environment", &validate.Rule{Regexp: envRegex})
		v.AddRule("AppId", &validate.Rule{Min: 1, Max: 160, Regexp: nameRegex})
		// support name or alias
		v.AddRule("ServiceName", &validate.Rule{Min: 1, Max: 160 + 1 + 128, Regexp: serviceNameForFindRegex})
	})
}

func ExistenceReqValidator() *validate.Validator {
	return existenceReqValidator.Init(func(v *validate.Validator) {
		v.AddRules(MicroServiceSearchKeyValidator().GetRules())
		v.AddRule("Version", &validate.Rule{Min: 1, Max: 64, Regexp: versionRegex})
	})
}

func GetServiceReqValidator() *validate.Validator {
	return getServiceReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", &validate.Rule{Min: 1, Max: 64, Regexp: serviceIDRegex})
	})
}

func CreateServiceReqValidator() *validate.Validator {
	return createServiceReqValidator.Init(func(v *validate.Validator) {
		max := int(quotasvc.SchemaQuota())

		var pathValidator validate.Validator
		pathValidator.AddRule("Path", &validate.Rule{Regexp: pathRegex})

		var frameworkValidator validate.Validator
		frameworkValidator.AddRule("Name", &validate.Rule{Max: 64, Regexp: nameRegex})
		frameworkValidator.AddRule("Version", &validate.Rule{Max: 64})

		var microServiceValidator validate.Validator
		microServiceValidator.AddRules(MicroServiceKeyValidator().GetRules())
		microServiceValidator.AddRule("AppId", &validate.Rule{Max: 160, Regexp: nameRegex})
		microServiceValidator.AddRule("Version", &validate.Rule{Max: 64, Regexp: versionRegex})
		microServiceValidator.AddRule("ServiceId", &validate.Rule{Max: 64, Regexp: serviceIDRegex})
		microServiceValidator.AddRule("Description", &validate.Rule{Max: 256})
		microServiceValidator.AddRule("Level", &validate.Rule{Regexp: levelRegex})
		microServiceValidator.AddRule("Status", &validate.Rule{Regexp: statusRegex})
		microServiceValidator.AddRule("Schemas", &validate.Rule{Max: max, Regexp: schemaIDRegex})
		microServiceValidator.AddSub("Paths", &pathValidator)
		microServiceValidator.AddRule("Alias", &validate.Rule{Max: 128, Regexp: aliasRegex})
		microServiceValidator.AddRule("RegisterBy", &validate.Rule{Max: 64, Regexp: registerByRegex})
		microServiceValidator.AddSub("Framework", &frameworkValidator)

		v.AddRule("Service", &validate.Rule{Min: 1})
		v.AddSub("Service", &microServiceValidator)
	})

}

func UpdateServicePropsReqValidator() *validate.Validator {
	return updateServicePropsReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
	})
}
