package gov_test

import (
	"github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPolicySchema_Validate(t *testing.T) {
	t.Run("validate simple schema", func(t *testing.T) {
		schema := &gov.PolicySchema{
			Attributes: map[string]*gov.PolicyAttribute{
				"allow": {
					Type: gov.ValueTypeBool,
				},
				"timeout": {
					Type: gov.ValueTypeString,
					Min:  1,
					Max:  5,
				},
				"name": {
					Type: gov.ValueTypeString,
				},
				"age": {
					Type: gov.ValueTypeInt,
				},
				"retry": {
					Type: gov.ValueTypeInt,
					Min:  1,
					Max:  5,
				},
			},
		}
		gov.RegisterPolicy("custom", schema)
		t.Run("given right content, should no error", func(t *testing.T) {
			err := schema.Validate([]byte(
				`allow: bool
timeout: 5s
name: a
age: 10
retry: 1`))
			assert.NoError(t, err)
		})
	})
}
