package validator

import (
	"regexp"

	ev "github.com/go-chassis/cari/env"

	"github.com/apache/servicecomb-service-center/pkg/validate"
)

var createEnvironmentReqValidator validate.Validator
var updateEnvironmentReqValidator validate.Validator
var envRegx, _ = regexp.Compile(`^\S*$`)
var desRegx, _ = regexp.Compile(`^\S*$`)

func ValidateCreateEnvironmentRequest(v *ev.CreateEnvironmentRequest) error {
	return CreateEnvironmentReqValidator().Validate(v)
}

func ValidateUpdateEnvironmentRequest(v *ev.UpdateEnvironmentRequest) error {
	return UpdateEnvironmentReqValidator().Validate(v)
}

func CreateEnvironmentReqValidator() *validate.Validator {
	return createEnvironmentReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("Environment", &validate.Rule{Min: 0, Max: 128, Regexp: envRegx})
		v.AddRule("Description", &validate.Rule{Min: 0, Max: 128, Regexp: desRegx})
	})
}

func UpdateEnvironmentReqValidator() *validate.Validator {
	return updateEnvironmentReqValidator.Init(func(v *validate.Validator) {
		v.AddRule("Environment", &validate.Rule{Min: 0, Max: 128, Regexp: envRegx})
		v.AddRule("Description", &validate.Rule{Min: 0, Max: 128, Regexp: desRegx})
	})
}
