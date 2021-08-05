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
	"math"
	"regexp"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/pkg/validate"
	"github.com/go-chassis/cari/discovery"
)

var (
	findInstanceReqValidator        validate.Validator
	batchFindInstanceReqValidator   validate.Validator
	getInstanceReqValidator         validate.Validator
	updateInstanceReqValidator      validate.Validator
	registerInstanceReqValidator    validate.Validator
	heartbeatReqValidator           validate.Validator
	updateInstancePropsReqValidator validate.Validator
)

var (
	instStatusRegex, _ = regexp.Compile("^(" + util.StringJoin([]string{
		discovery.MSI_UP, discovery.MSI_DOWN, discovery.MSI_STARTING, discovery.MSI_TESTING, discovery.MSI_OUTOFSERVICE}, "|") + ")?$")
	updateInstStatusRegex, _ = regexp.Compile("^(" + util.StringJoin([]string{
		discovery.MSI_UP, discovery.MSI_DOWN, discovery.MSI_STARTING, discovery.MSI_TESTING, discovery.MSI_OUTOFSERVICE}, "|") + ")$")
	hbModeRegex, _               = regexp.Compile(`^(push|pull)$`)
	urlRegex, _                  = regexp.Compile(`^\S*$`)
	epRegex, _                   = regexp.Compile(`\S+`)
	simpleNameAllowEmptyRegex, _ = regexp.Compile(`^[A-Za-z0-9_.-]*$`)
	simpleNameRegex, _           = regexp.Compile(`^[A-Za-z0-9_.-]+$`)
	regionRegex, _               = regexp.Compile(`^[A-Za-z0-9_.-]+$`)
)

func FindInstanceReqValidator() *validate.Validator {
	return findInstanceReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ConsumerServiceId", GetInstanceReqValidator().GetRule("ConsumerServiceId"))
		v.AddRules(MicroServiceSearchKeyValidator().GetRules())
		v.AddRule("Tags", UpdateTagReqValidator().GetRule("Key"))
	})
}

func BatchFindInstanceReqValidator() *validate.Validator {
	return batchFindInstanceReqValidator.Init(func(v *validate.Validator) {
		var findServiceValidator validate.Validator
		findServiceValidator.AddRule("Service", &validate.Rule{Min: 1})
		findServiceValidator.AddSub("Service", MicroServiceSearchKeyValidator())
		var findInstanceValidator validate.Validator
		findInstanceValidator.AddRule("Instance", &validate.Rule{Min: 1})
		findInstanceValidator.AddSub("Instance", HeartbeatReqValidator())
		v.AddRule("ConsumerServiceId", GetInstanceReqValidator().GetRule("ConsumerServiceId"))
		v.AddSub("Services", &findServiceValidator)
		v.AddSub("Instances", &findInstanceValidator)
	})
}

func GetInstanceReqValidator() *validate.Validator {
	return getInstanceReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ConsumerServiceId", &validate.Rule{Max: 64, Regexp: serviceIDRegex})
		v.AddRule("ProviderServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("ProviderInstanceId", HeartbeatReqValidator().GetRule("InstanceId"))
		v.AddRule("Tags", UpdateTagReqValidator().GetRule("Key"))
	})
}

func HeartbeatReqValidator() *validate.Validator {
	return heartbeatReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("InstanceId", &validate.Rule{Min: 1, Max: 64, Regexp: simpleNameAllowEmptyRegex})
	})
}

func UpdateInstanceReqValidator() *validate.Validator {
	return updateInstanceReqValidator.Init(func(v *validate.Validator) {
		v.AddRules(heartbeatReqValidator.GetRules())
		v.AddRule("Status", &validate.Rule{Regexp: updateInstStatusRegex})
	})
}

func UpdateInstancePropsReqValidator() *validate.Validator {
	return updateInstancePropsReqValidator.Init(func(v *validate.Validator) {
		v.AddRules(heartbeatReqValidator.GetRules())
	})
}

func RegisterInstanceReqValidator() *validate.Validator {
	return registerInstanceReqValidator.Init(func(v *validate.Validator) {
		var healthCheckInfoValidator validate.Validator
		healthCheckInfoValidator.AddRule("Mode", &validate.Rule{Regexp: hbModeRegex})
		healthCheckInfoValidator.AddRule("Port", &validate.Rule{Max: math.MaxUint16, Min: 0})
		healthCheckInfoValidator.AddRule("Times", &validate.Rule{Max: math.MaxInt32})
		healthCheckInfoValidator.AddRule("Interval", &validate.Rule{Min: 1, Max: math.MaxInt32})
		healthCheckInfoValidator.AddRule("Url", &validate.Rule{Regexp: urlRegex})

		var dataCenterInfoValidator validate.Validator
		dataCenterInfoValidator.AddRule("Name", &validate.Rule{Min: 1, Max: 128, Regexp: simpleNameRegex})
		dataCenterInfoValidator.AddRule("Region", &validate.Rule{Min: 1, Max: 128, Regexp: regionRegex})
		dataCenterInfoValidator.AddRule("AvailableZone", &validate.Rule{Min: 1, Max: 128, Regexp: regionRegex})

		var microServiceInstanceValidator validate.Validator
		microServiceInstanceValidator.AddRule("InstanceId", &validate.Rule{Max: 64, Regexp: simpleNameAllowEmptyRegex})
		microServiceInstanceValidator.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		// allow empty endpoint register for client only
		microServiceInstanceValidator.AddRule("Endpoints", &validate.Rule{Regexp: epRegex})
		microServiceInstanceValidator.AddRule("HostName", &validate.Rule{Max: 64, Min: 1, Regexp: epRegex})
		microServiceInstanceValidator.AddSub("HealthCheck", &healthCheckInfoValidator)
		microServiceInstanceValidator.AddRule("Status", &validate.Rule{Regexp: instStatusRegex})
		microServiceInstanceValidator.AddSub("DataCenterInfo", &dataCenterInfoValidator)

		v.AddRule("Instance", &validate.Rule{Min: 1})
		v.AddSub("Instance", &microServiceInstanceValidator)
	})
}
