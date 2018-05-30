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
	"github.com/apache/incubator-servicecomb-service-center/pkg/validate"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/quota"
	"regexp"
)

var (
	updateRuleReqValidator validate.Validator
	addRulesReqValidator   validate.Validator
)

var (
	ruleRegex, _     = regexp.Compile(`^(WHITE|BLACK)$`)
	ruleAttrRegex, _ = regexp.Compile(`((^tag_[a-zA-Z][a-zA-Z0-9_\-.]{0,63}$)|(^ServiceId$)|(^AppId$)|(^ServiceName$)|(^Version$)|(^Description$)|(^Level$)|(^Status$))`)
)

func UpdateRuleReqValidator() *validate.Validator {
	var ruleValidator validate.Validator
	ruleValidator.AddRule("RuleType", &validate.ValidateRule{Regexp: ruleRegex})
	ruleValidator.AddRule("Attribute", &validate.ValidateRule{Regexp: ruleAttrRegex})
	ruleValidator.AddRule("Pattern", &validate.ValidateRule{Min: 1, Max: 64})
	ruleValidator.AddRule("Description", CreateServiceReqValidator().GetSub("Service").GetRule("Description"))

	return updateRuleReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("RuleId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddSub("Rule", &ruleValidator)
	})
}

func AddRulesReqValidator() *validate.Validator {
	return addRulesReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Rules", &validate.ValidateRule{Min: 1, Max: quota.DefaultRuleQuota})
		v.AddSub("Rules", UpdateRuleReqValidator().GetSub("Rule"))
	})
}
