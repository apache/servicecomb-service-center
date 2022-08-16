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

// Package grc include API of governance(grc is the abbreviation of governance)
package grc

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

type ValueType string

// policySchemas saves policy kind and schema
var policySchemas = make(map[string]*spec.Schema)

// RegisterPolicySchema register a contract of one kind of policy
// this API is not thread safe, only use it during sc init
func RegisterPolicySchema(kind string, schema *spec.Schema) {
	policySchemas[kind] = schema
	log.Info("register policy schema: " + kind)
}

// ValidatePolicySpec validates spec attributes
// it first us legacy mechanism Validate 3 kinds of policy.
// then it use new mechanism to validate all of other policy
func ValidatePolicySpec(kind string, spec interface{}) error {
	// TODO this legacy API should be removed. after 3 kinds of policy schema is registered by "RegisterPolicySchema"
	err := Validate(kind, spec)
	if err != nil {
		return err
	}
	schema, ok := policySchemas[kind]
	if !ok {
		log.Warn(fmt.Sprintf("can not recognize policy %s", kind))
		return &ErrIllegalItem{"not support kind yet", kind}
	}
	validator := validate.NewSchemaValidator(schema, nil, "", strfmt.Default)
	errs := validator.Validate(spec).Errors
	if len(errs) != 0 {
		var str []string
		for i, err := range errs {
			if i != 0 {
				str = append(str, ";", err.Error())
			} else {
				str = append(str, err.Error())
			}
		}
		return errors.New(strings.Join(str, ";"))
	}
	return nil
}
