package gov_test

import (
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"github.com/apache/servicecomb-service-center/server/service/gov"
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
		gov.RegisterPolicy("custom", schema)
		t.Run("given right content, should no error", func(t *testing.T) {
			err := gov.ValidateSpec("custom", map[string]interface{}{
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
			err := gov.ValidateSpec("custom", map[string]interface{}{
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
			err := gov.ValidateSpec("custom", map[string]interface{}{
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
