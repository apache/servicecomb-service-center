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

package gov

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

//policies saves kind and policy schemas
var policies = make(map[string]*spec.Schema)

//RegisterPolicy register a contract of one kind of policy
//this API is not thread safe, only use it during sc init
func RegisterPolicy(kind string, schema *spec.Schema) {
	policies[kind] = schema
}

//ValidateSpec validates spec attributes
func ValidateSpec(kind string, spec interface{}) error {
	schema, ok := policies[kind]
	if !ok {
		log.Warn(fmt.Sprintf("can not recognize %s", kind))
		return nil
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
