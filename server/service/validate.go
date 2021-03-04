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
	"errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/validate"
	"reflect"
	"regexp"
)

var createAccountValidator = &validate.Validator{}
var changePWDValidator = &validate.Validator{}
var accountLoginValidator = &validate.Validator{}
var pwdValidator = &validate.Rule{Regexp: &validate.PasswordChecker{}, Hide: true}

func init() {
	var roleRegex, _ = regexp.Compile(`^$|^(admin|developer)$`)
	var accountRegex, _ = regexp.Compile(`^[a-zA-Z]\w{3,15}$`)
	createAccountValidator.AddRule("Name", &validate.Rule{Regexp: accountRegex})
	createAccountValidator.AddRule("Role", &validate.Rule{Regexp: roleRegex})
	createAccountValidator.AddRule("Password", pwdValidator)

	changePWDValidator.AddRule("Password", pwdValidator)
	changePWDValidator.AddRule("Name", &validate.Rule{Regexp: accountRegex})

	accountLoginValidator.AddRule("TokenExpirationTime", &validate.Rule{Regexp: &validate.TokenExpirationTimeChecker{}})
}

func Validate(v interface{}) error {
	err := baseCheck(v)
	if err != nil {
		return err
	}
	switch t := v.(type) {
	case *pb.CreateServiceRequest:
		return CreateServiceReqValidator().Validate(v)
	case *pb.GetServiceRequest,
		*pb.DeleteServiceRequest,
		*pb.GetDependenciesRequest:
		return GetServiceReqValidator().Validate(v)
	case *pb.UpdateServicePropsRequest:
		return UpdateServicePropsReqValidator().Validate(v)
	case *pb.CreateDependenciesRequest:
		return CreateDependenciesReqValidator().Validate(v)
	case *pb.AddDependenciesRequest:
		return AddDependenciesReqValidator().Validate(v)
	case *pb.GetServiceTagsRequest:
		return GetTagsReqValidator().Validate(v)
	case *pb.AddServiceTagsRequest:
		return AddTagsReqValidator().Validate(v)
	case *pb.UpdateServiceTagRequest:
		return UpdateTagReqValidator().Validate(v)
	case *pb.DeleteServiceTagsRequest:
		return DeleteTagReqValidator().Validate(v)
	case *pb.GetAllSchemaRequest:
		return GetSchemaReqValidator().Validate(v)
	case *pb.GetSchemaRequest,
		*pb.DeleteSchemaRequest:
		return GetSchemaReqValidator().Validate(v)
	case *pb.ModifySchemaRequest:
		return ModifySchemaReqValidator().Validate(v)
	case *pb.ModifySchemasRequest:
		return ModifySchemasReqValidator().Validate(v)
	case *pb.GetOneInstanceRequest,
		*pb.GetInstancesRequest:
		return GetInstanceReqValidator().Validate(v)
	case *pb.UpdateInstanceStatusRequest:
		return UpdateInstanceReqValidator().Validate(v)
	case *pb.RegisterInstanceRequest:
		return RegisterInstanceReqValidator().Validate(v)
	case *pb.FindInstancesRequest:
		return FindInstanceReqValidator().Validate(v)
	case *pb.BatchFindInstancesRequest:
		return BatchFindInstanceReqValidator().Validate(v)
	case *pb.HeartbeatRequest, *pb.UnregisterInstanceRequest:
		return HeartbeatReqValidator().Validate(v)
	case *pb.UpdateInstancePropsRequest:
		return UpdateInstancePropsReqValidator().Validate(v)
	case *pb.GetServiceRulesRequest:
		return GetRulesReqValidator().Validate(v)
	case *pb.AddServiceRulesRequest:
		return AddRulesReqValidator().Validate(v)
	case *pb.UpdateServiceRuleRequest:
		return UpdateRuleReqValidator().Validate(v)
	case *pb.DeleteServiceRulesRequest:
		return DeleteRulesReqValidator().Validate(v)
	case *pb.GetAppsRequest:
		return MicroServiceKeyValidator().Validate(v)
	default:
		log.Warnf("No validator for %T.", t)
		return nil
	}
}

func baseCheck(v interface{}) error {
	if v == nil {
		return errors.New("data is nil")
	}
	sv := reflect.ValueOf(v)
	if sv.Kind() == reflect.Ptr && sv.IsNil() {
		return errors.New("pointer is nil")
	}
	return nil
}
func ValidateCreateAccount(a *rbacframe.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return createAccountValidator.Validate(a)
}
func ValidateAccountLogin(a *rbacframe.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return accountLoginValidator.Validate(a)
}
func ValidateChangePWD(a *rbacframe.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return changePWDValidator.Validate(a)
}
