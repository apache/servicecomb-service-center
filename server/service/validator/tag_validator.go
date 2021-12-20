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

	"github.com/apache/servicecomb-service-center/pkg/validate"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
)

var (
	getTagsReqValidator   validate.Validator
	addTagsReqValidator   validate.Validator
	updateTagReqValidator validate.Validator
	deleteTagReqValidator validate.Validator
)

var (
	tagRegex, _ = regexp.Compile(`^[a-zA-Z][a-zA-Z0-9_\-.]{0,63}$`)
)

func GetTagsReqValidator() *validate.Validator {
	return getTagsReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
	})
}

func AddTagsReqValidator() *validate.Validator {
	return addTagsReqValidator.Init(func(v *validate.Validator) {
		max := int(quotasvc.TagQuota())
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Tags", &validate.Rule{Max: max, Regexp: tagRegex})
	})
}

func UpdateTagReqValidator() *validate.Validator {
	return updateTagReqValidator.Init(func(v *validate.Validator) {
		tagRule := &validate.Rule{Regexp: tagRegex}
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Key", tagRule)
		v.AddRule("Value", tagRule)
	})
}

func DeleteTagReqValidator() *validate.Validator {
	return deleteTagReqValidator.Init(func(v *validate.Validator) {
		max := int(quotasvc.TagQuota())
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Keys", &validate.Rule{Min: 1, Max: max, Regexp: tagRegex})
	})
}
