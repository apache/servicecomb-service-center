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

package policy

import (
	"github.com/apache/servicecomb-service-center/server/service/grc"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

var (
	notEmpty         = int64(1)
	notNegative      = float64(0)
	matchGroupSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"matches", "alias"},
			Properties: map[string]spec.Schema{
				"alias": {
					SchemaProps: spec.SchemaProps{
						Type:      []string{"string"},
						MinLength: &notEmpty,
					},
				},
				"matches": {
					SchemaProps: spec.SchemaProps{
						Type:     []string{"array"},
						MinItems: &notEmpty,
						Items: &spec.SchemaOrArray{
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type:     []string{"object"},
									Required: []string{"name"},
									Properties: map[string]spec.Schema{
										"name": {
											SchemaProps: spec.SchemaProps{
												Type:      []string{"string"},
												MinLength: &notEmpty,
											},
										},
										"apiPath": {
											SchemaProps: spec.SchemaProps{
												Type:      []string{"string"},
												MinLength: &notEmpty,
											},
										},
										"method": {
											SchemaProps: spec.SchemaProps{
												Type:     []string{"array"},
												MinItems: &notEmpty,
												Items: &spec.SchemaOrArray{
													Schema: &spec.Schema{
														SchemaProps: spec.SchemaProps{
															Type: []string{"string"},
															Enum: []interface{}{"GET", "POST", "DELETE", "PUT", "PATCH"},
														},
													},
												},
											},
										},
										"headers": {
											SchemaProps: spec.SchemaProps{
												Type:          []string{"object"},
												MinProperties: &notEmpty,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}}
	retrySchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:          []string{"object"},
			MinProperties: &notEmpty,
			Properties: map[string]spec.Schema{
				"maxAttempts": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"integer"},
						Minimum: &notNegative,
					},
				},
				"retryOnSame": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"integer"},
						Minimum: &notNegative,
					},
				},
			},
		},
	}
	rateLimitingSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"rate"},
			Properties: map[string]spec.Schema{
				"rate": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"number"},
						Minimum: &notNegative,
					},
				},
			},
		},
	}
	loadbalanceSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"rule"},
			Properties: map[string]spec.Schema{
				"rule": {
					SchemaProps: spec.SchemaProps{
						Type:      []string{"string"},
						MinLength: &notEmpty,
					},
				},
			},
		}}
	circuitBreakerSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"minimumNumberOfCalls"},
			Properties: map[string]spec.Schema{
				"minimumNumberOfCalls": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"integer"},
						Minimum: &notNegative,
					},
				},
			},
		}}
	instanceIsolationSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"minimumNumberOfCalls"},
			Properties: map[string]spec.Schema{
				"minimumNumberOfCalls": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"integer"},
						Minimum: &notNegative,
					},
				},
			},
		}}
	faultInjectionSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"percentage"},
			Properties: map[string]spec.Schema{
				"percentage": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"number"},
						Minimum: &notNegative,
					},
				},
			},
		}}
	bulkheadSchema = &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"maxConcurrentCalls"},
			Properties: map[string]spec.Schema{
				"maxConcurrentCalls": {
					SchemaProps: spec.SchemaProps{
						Type:    []string{"integer"},
						Minimum: &notNegative,
					},
				},
			},
		}}
)

func init() {
	grc.RegisterPolicySchema("match-group", matchGroupSchema)
	grc.RegisterPolicySchema("retry", retrySchema)
	grc.RegisterPolicySchema("rate-limiting", rateLimitingSchema)
	grc.RegisterPolicySchema("loadbalance", loadbalanceSchema)
	grc.RegisterPolicySchema("circuit-breaker", circuitBreakerSchema)
	grc.RegisterPolicySchema("instance-isolation", instanceIsolationSchema)
	grc.RegisterPolicySchema("fault-injection", faultInjectionSchema)
	grc.RegisterPolicySchema("bulkhead", bulkheadSchema)
}
