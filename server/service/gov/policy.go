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
