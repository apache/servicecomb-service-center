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
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

var (
	getRulesReqValidator    validate.Validator
	updateRuleReqValidator  validate.Validator
	addRulesReqValidator    validate.Validator
	deleteRulesReqValidator validate.Validator
)

var (
	ruleRegex, _     = regexp.Compile(`^(WHITE|BLACK)$`)
	ruleAttrRegex, _ = regexp.Compile(`((^tag_[a-zA-Z][a-zA-Z0-9_\-.]{0,63}$)|(^ServiceId$)|(^AppId$)|(^ServiceName$)|(^Version$)|(^Description$)|(^Level$)|(^Status$))`)
)

func GetRulesReqValidator() *validate.Validator {
	return getRulesReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
	})
}

func UpdateRuleReqValidator() *validate.Validator {
	return updateRuleReqValidator.Init(func(v *validate.Validator) {
		var ruleValidator validate.Validator
		ruleValidator.AddRule("RuleType", &validate.Rule{Regexp: ruleRegex})
		ruleValidator.AddRule("Attribute", &validate.Rule{Regexp: ruleAttrRegex})
		ruleValidator.AddRule("Pattern", &validate.Rule{Min: 1, Max: 64})
		ruleValidator.AddRule("Description", CreateServiceReqValidator().GetSub("Service").GetRule("Description"))

		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("RuleId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddSub("Rule", &ruleValidator)
	})
}

func AddRulesReqValidator() *validate.Validator {
	return addRulesReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Rules", &validate.Rule{Min: 1, Max: quota.DefaultRuleQuota})
		v.AddSub("Rules", UpdateRuleReqValidator().GetSub("Rule"))
	})
}

func DeleteRulesReqValidator() *validate.Validator {
	return deleteRulesReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("RuleIds", &validate.Rule{Min: 1, Max: quota.DefaultRuleQuota})
	})
}
