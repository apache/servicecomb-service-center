package policy

import (
	"github.com/apache/servicecomb-service-center/server/service/grc"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func init() {
	grc.RegisterPolicySchema("loadbalancer", &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:     []string{"object"},
			Required: []string{"rule"},
			Properties: map[string]spec.Schema{
				"rule": {
					SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
			},
		}})
	grc.RegisterPolicySchema("circuitBreaker", &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"minimumNumberOfCalls": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
			},
		}})
	grc.RegisterPolicySchema("instanceIsolation", &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"minimumNumberOfCalls": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
			},
		}})
	grc.RegisterPolicySchema("faultInjection", &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"percentage": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
			},
		}})
	grc.RegisterPolicySchema("bulkhead", &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"maxConcurrentCalls": {
					SchemaProps: spec.SchemaProps{Type: []string{"integer"}}},
			},
		}})

}
