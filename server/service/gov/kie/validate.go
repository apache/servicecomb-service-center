/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package kie

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/server/service/gov"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

type Validator struct {
}

type ErrIllegalItem struct {
	err string
	val interface{}
}

var (
	methodSet            []interface{}
	slidingWindowTypeSet []interface{}
	matchGroupParamsSet  []spec.Schema
	minValue             float64
	maxValue             float64
	minLength            int64
)

func (e *ErrIllegalItem) Error() string {
	return fmt.Sprintf("illegal item : %v , msg: %s", e.val, e.err)
}

func (d *Validator) Validate(kind string, spec map[string]interface{}) error {
	switch kind {
	case "match-group":
		return matchValidate(spec)
	case "retry":
		return retryValidate(spec)
	case "rate-limiting":
		return rateLimitingValidate(spec)
	case "circuit-breaker":
		return circuitBreakerValidate(spec)
	case "bulkhead":
		return bulkheadValidate(spec)
	case "loadbalancer":
		return nil
	default:
		return &ErrIllegalItem{"not support kind yet", kind}
	}
	return nil
}

func matchValidate(spec map[string]interface{}) error {
	err := gov.ValidateSpec("MatchGroup", spec)
	if err != nil {
		return err
	}
	return nil
}

func retryValidate(spec map[string]interface{}) error {
	err := policyValidate(spec)
	if err != nil {
		return err
	}
	return nil
}

func rateLimitingValidate(spec map[string]interface{}) error {
	err := policyValidate(spec)
	if err != nil {
		return err
	}
	return nil
}

func bulkheadValidate(spec map[string]interface{}) error {
	err := gov.ValidateSpec("bulkhead", spec)
	if err != nil {
		return err
	}
	return nil
}

func circuitBreakerValidate(spec map[string]interface{}) error {
	err := gov.ValidateSpec("circuitBreaker", spec)
	if err != nil {
		return err
	}
	return nil
}

func policyValidate(spec map[string]interface{}) error {
	err := gov.ValidateSpec("policy", spec)
	if err != nil {
		return err
	}
	return nil
}

func apiPathSchema() spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"apiPath"},
			Properties: map[string]spec.Schema{
				"apiPath": {
					SchemaProps: spec.SchemaProps{Type: []string{"object"}}},
			},
		}}
}

func headersSchema() spec.Schema {
	return spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"headers"},
			Properties: map[string]spec.Schema{
				"headers": {
					SchemaProps: spec.SchemaProps{Type: []string{"object"}}},
			},
		}}
}

func matchGroupSchema() *spec.Schema {
	methodSchema := spec.SchemaOrArray{
		Schema: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
				Enum: methodSet,
			}},
		Schemas: nil,
	}

	matchSchema := spec.SchemaOrArray{
		Schema: &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:     []string{"object"},
				AnyOf:    matchGroupParamsSet,
				Required: []string{"name", "method"},
				Properties: map[string]spec.Schema{
					"name": {
						SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
					"method": {
						SchemaProps: spec.SchemaProps{Type: []string{"array"}, Items: &methodSchema}},
				},
			}},
		Schemas: nil,
	}

	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"alias", Matches},
			Properties: map[string]spec.Schema{
				"alias": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
				Matches: {
					SchemaProps: spec.SchemaProps{Type: []string{"array"}, Items: &matchSchema}},
			},
		}}
}

func policySchema() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				Rules: {
					SchemaProps: spec.SchemaProps{Type: []string{"object"}, PatternProperties: map[string]spec.Schema{
						"match": {
							SchemaProps: spec.SchemaProps{Type: []string{"string"}, Nullable: false, MinLength: &minLength}},
					}}},
			},
		}}
}

func bulkheadSchema() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"name", "maxConcurrentCalls", "maxWaitDuration"},
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}, Nullable: false, MinLength: &minLength}},
				"maxConcurrentCalls": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}, Minimum: &minValue}},
				"maxWaitDuration": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}, Nullable: false, MinLength: &minLength}},
			},
		}}
}

func circuitBreakerSchema() *spec.Schema {
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Required: []string{"failureRateThreshold", "slowCallRateThreshold", "slowCallDurationThreshold",
				"minimumNumberOfCalls", "slidingWindowType", "slidingWindowSize"},
			Properties: map[string]spec.Schema{
				"failureRateThreshold": {
					SchemaProps: spec.SchemaProps{Type: []string{"number"}, Minimum: &minValue, ExclusiveMinimum: true, Maximum: &maxValue}},
				"slowCallRateThreshold": {
					SchemaProps: spec.SchemaProps{Type: []string{"number"}, Minimum: &minValue, ExclusiveMinimum: true, Maximum: &maxValue}},
				"slowCallDurationThreshold": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}, Nullable: false, MinLength: &minLength}},
				"minimumNumberOfCalls": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}, Minimum: &minValue, ExclusiveMinimum: true}},
				"slidingWindowType": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}, Enum: slidingWindowTypeSet}},
				"slidingWindowSize": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}, Nullable: false, MinLength: &minLength}},
			},
		}}
}

func init() {
	minValue = 0
	maxValue = 100
	minLength = 1
	methodSet = make([]interface{}, 0)
	slidingWindowTypeSet = make([]interface{}, 0)
	matchGroupParamsSet = make([]spec.Schema, 0)
	methodSet = append(methodSet, "GET", "POST", "DELETE", "PUT", "PATCH")
	slidingWindowTypeSet = append(slidingWindowTypeSet, "count", "time")
	matchGroupParamsSet = append(matchGroupParamsSet, apiPathSchema(), headersSchema())

	gov.RegisterPolicy("MatchGroup", matchGroupSchema())
	gov.RegisterPolicy("bulkhead", bulkheadSchema())
	gov.RegisterPolicy("circuitBreaker", circuitBreakerSchema())
	gov.RegisterPolicy("policy", policySchema())
}
