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
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type ValueType string

var PolicyNames []string

// policySchemas saves policy kind and schema
var policySchemas = make(map[string]*spec.Schema)

// RegisterPolicySchema register a contract of one kind of policy
// this API is not thread safe, only use it during sc init
func RegisterPolicySchema(kind string, schema *spec.Schema) {
	policySchemas[kind] = schema
	log.Info("register policy schema: " + kind)
}

// ValidatePolicySpec validates spec attributes
func ValidatePolicySpec(kind string, spec interface{}) error {
	schema, ok := policySchemas[kind]
	if !ok {
		log.Warn(fmt.Sprintf("can not recognize policy %s", kind))
		return fmt.Errorf("not support kind[%s] yet", kind)
	}
	mathSpec := &gov.MatchSpec{}
	err := convertInterfaceToType(spec, mathSpec)
	if err != nil {
		return err
	}
	err = checkMathSpec(mathSpec)
	if err != nil {
		return err
	}
	schemaValidator := validate.NewSchemaValidator(schema, nil, kind, strfmt.Default)
	errs := schemaValidator.Validate(spec).Errors
	if len(errs) != 0 {
		var str []string
		for _, err := range errs {
			str = append(str, err.Error())
		}
		return fmt.Errorf("illegal policy[%s] spec, msg: %s", kind, strings.Join(str, "; "))
	}

	return nil
}

func checkMathSpec(matchSpec *gov.MatchSpec) error {
	if matchSpec == nil || matchSpec.MatchPolicies == nil {
		return fmt.Errorf("match-group.spec or match-group.spec.matches can not  null")
	}
	for i, policy := range matchSpec.MatchPolicies {
		if len(policy.APIPaths) == 0 {
			return fmt.Errorf("match-group.spec.matches[%d].apiPath can not be empty", i)
		}
		for key, path := range policy.APIPaths {
			if len(path) == 0 {
				return fmt.Errorf("match-group.spec.matches[%d].apiPath[%s] path can not be empty", i, key)
			}
		}
		if policy.Methods == nil || len(policy.Methods) == 0 {
			return fmt.Errorf("match-group.spec.matches[%d].method can not be empty", i)
		}
	}
	return nil
}

// convertInterfaceToType 使用 JSON 序列化和反序列化将 interface{} 类型的值转换为目标类型
func convertInterfaceToType(src interface{}, target interface{}) error {
	jsonBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %w", err)
	}

	return nil
}
