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
	"regexp"

	"github.com/apache/servicecomb-service-center/pkg/validate"
)

var (
	addDependenciesReqValidator       validate.Validator
	overwriteDependenciesReqValidator validate.Validator
)

var (
	nameFuzzyRegex, _         = regexp.Compile(`^[a-zA-Z0-9]*$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$|^\*$`)
	versionAllowEmptyRegex, _ = regexp.Compile(`^(^\d+(\.\d+){0,2}\+?$|^\d+(\.\d+){0,2}-\d+(\.\d+){0,2}$|^latest$)?$`)
)

func defaultDependencyValidator() *validate.Validator {
	appIDRule := *(MicroServiceKeyValidator().GetRule("AppId"))
	appIDRule.Min = 0
	serviceNameRule := *(MicroServiceKeyValidator().GetRule("ServiceName"))
	serviceNameRule.Regexp = nameFuzzyRegex
	versionRule := &validate.Rule{Max: 128, Regexp: &validate.VersionRegexp{Fuzzy: true, Regex: versionAllowEmptyRegex}}

	var (
		consumerMsValidator validate.Validator
		providerMsValidator validate.Validator
	)
	consumerMsValidator.AddRules(MicroServiceKeyValidator().GetRules())

	providerMsValidator.AddRules(MicroServiceKeyValidator().GetRules())
	providerMsValidator.AddRule("AppId", &appIDRule)
	providerMsValidator.AddRule("ServiceName", &serviceNameRule)
	providerMsValidator.AddRule("Version", versionRule)

	var dependenciesValidator validate.Validator
	dependenciesValidator.AddRule("Consumer", &validate.Rule{Min: 1})
	dependenciesValidator.AddSub("Consumer", &consumerMsValidator)
	dependenciesValidator.AddSub("Providers", &providerMsValidator)

	return &dependenciesValidator
}

func AddDependenciesReqValidator() *validate.Validator {
	return addDependenciesReqValidator.Init(func(v *validate.Validator) {
		dep := defaultDependencyValidator()
		dep.AddRule("Providers", &validate.Rule{Min: 1})
		v.AddRule("Dependencies", &validate.Rule{Min: 1, Max: 100})
		v.AddSub("Dependencies", dep)
	})
}

func CreateDependenciesReqValidator() *validate.Validator {
	return overwriteDependenciesReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("Dependencies", &validate.Rule{Min: 1, Max: 100})
		v.AddSub("Dependencies", defaultDependencyValidator())
	})
}
