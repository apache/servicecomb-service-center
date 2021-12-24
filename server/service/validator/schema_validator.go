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
	pb "github.com/go-chassis/cari/discovery"
)

var (
	getSchemaReqValidator     validate.Validator
	modifySchemasReqValidator validate.Validator
	modifySchemaReqValidator  validate.Validator
)

var (
	schemaIDUnlimitedRegex, _ = regexp.Compile(`^[a-zA-Z0-9]+$|^[a-zA-Z0-9][a-zA-Z0-9_\-.]*[a-zA-Z0-9]$`)
	schemaSummaryRegex, _     = regexp.Compile(`^[a-zA-Z0-9]*$`)
)

func GetSchemaReqValidator() *validate.Validator {
	return getSchemaReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("SchemaId", &validate.Rule{Min: 1, Max: 160, Regexp: schemaIDUnlimitedRegex})
	})
}

func ModifySchemasReqValidator() *validate.Validator {
	return modifySchemasReqValidator.Init(func(v *validate.Validator) {
		max := int(quotasvc.SchemaQuota())

		var subSchemaValidator validate.Validator
		subSchemaValidator.AddRule("SchemaId", GetSchemaReqValidator().GetRule("SchemaId"))
		subSchemaValidator.AddRule("Summary", &validate.Rule{Min: 1, Max: 128, Regexp: schemaSummaryRegex})
		subSchemaValidator.AddRule("Schema", &validate.Rule{Min: 1})

		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		v.AddRule("Schemas", &validate.Rule{Min: 1, Max: max})
		v.AddSub("Schemas", &subSchemaValidator)
	})
}

func ModifySchemaReqValidator() *validate.Validator {
	return modifySchemaReqValidator.Init(func(v *validate.Validator) {
		v.AddRules(ModifySchemasReqValidator().GetSub("Schemas").GetRules())
		v.AddRule("ServiceId", GetServiceReqValidator().GetRule("ServiceId"))
		// forward compatibility: allow empty
		v.AddRule("Summary", &validate.Rule{Max: 128, Regexp: schemaSummaryRegex})
	})
}

func ValidateGetSchema(request *pb.GetSchemaRequest) error {
	err := baseCheck(request)
	if err != nil {
		return err
	}
	return GetSchemaReqValidator().Validate(request)
}
func ValidateListSchema(request *pb.GetAllSchemaRequest) error {
	err := baseCheck(request)
	if err != nil {
		return err
	}
	return GetSchemaReqValidator().Validate(request)
}
func ValidatePutSchema(request *pb.ModifySchemaRequest) error {
	err := baseCheck(request)
	if err != nil {
		return err
	}
	return ModifySchemaReqValidator().Validate(request)
}
func ValidatePutSchemas(request *pb.ModifySchemasRequest) error {
	err := baseCheck(request)
	if err != nil {
		return err
	}
	return ModifySchemasReqValidator().Validate(request)
}
func ValidateDeleteSchema(request *pb.DeleteSchemaRequest) error {
	err := baseCheck(request)
	if err != nil {
		return err
	}
	return GetSchemaReqValidator().Validate(request)
}
