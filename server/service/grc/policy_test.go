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

package grc_test

import (
	"github.com/apache/servicecomb-service-center/server/service/grc"
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"github.com/stretchr/testify/assert"
)

func TestPolicySchema_Validate(t *testing.T) {
	t.Run("validate simple schema", func(t *testing.T) {
		schema := &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:     []string{"object"},
				Required: []string{"name"},
				Properties: map[string]spec.Schema{
					"allow": {
						SchemaProps: spec.SchemaProps{Type: []string{"boolean"}}},
					"name": {
						SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
					"timeout": {
						SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
					"age": {
						SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
					"retry": {
						SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
				},
			}}
		grc.RegisterPolicySchema("custom", schema)
		t.Run("given right content, should no error", func(t *testing.T) {
			err := grc.ValidatePolicySpec("custom", map[string]interface{}{
				"allow":   true,
				"timeout": "5s",
				"name":    "a",
				"age":     10,
				"retry":   1,
			})
			t.Log(err)
			assert.NoError(t, err)
		})
		t.Run("given value with wrong type, should return error", func(t *testing.T) {
			err := grc.ValidatePolicySpec("custom", map[string]interface{}{
				"allow":   "str",
				"timeout": "5s",
				"name":    "a",
				"age":     10,
				"retry":   1,
			})
			t.Log(err)
			assert.Error(t, err)
		})
		t.Run("do not give required value, should return error", func(t *testing.T) {
			err := grc.ValidatePolicySpec("custom", map[string]interface{}{
				"allow":   true,
				"timeout": "5s",
				"age":     10,
				"retry":   1,
			})
			t.Log(err)
			assert.Error(t, err)
		})
	})
}
