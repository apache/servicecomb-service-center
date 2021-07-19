package gov

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"gopkg.in/yaml.v2"
)

type ValueType string

var (
	MsgRequired    = "%s is required"
	MsgTooSmall    = "%s should be bigger than %d"
	MsgTooBig      = "%s should be smaller than %d"
	MsgTypeInt     = "%s must be a number"
	MsgTypeBool    = "%s must be a bool"
	MsgSkip        = "skip checking %s"
	MsgUnknownType = "%s is unknown type"
)

//policies saves kind and policy schemas
var policies = make(map[string]*PolicySchema, 0)

const (
	ValueTypeInt    ValueType = "int"
	ValueTypeString ValueType = "string"
	ValueTypeMap    ValueType = "map" // TODO not supported
	ValueTypeList   ValueType = "list"
	ValueTypeBool   ValueType = "bool"
)

type validate func(k string, v interface{}, attribute *PolicyAttribute) error

type PolicySchema struct {
	Attributes map[string]*PolicyAttribute
}
type PolicyAttribute struct {
	Type            ValueType
	Required        bool
	Min             int
	Max             int
	NestedAttribute map[string]*PolicyAttribute
}

func (schema *PolicySchema) Validate(spec []byte) error {
	m := make(map[string]interface{})
	if err := yaml.Unmarshal(spec, m); err != nil {
		return err
	}
	for k, v := range m {
		a, ok := schema.Attributes[k]
		if !ok {
			log.Debug(fmt.Sprintf(MsgSkip, k))
			continue
		}
		err := validateAny(k, v, a)
		if err != nil {
			return err
		}
	}
	return nil
}

//RegisterPolicy register a contract of one kind of policy
//this API is not thread safe, only use it during sc init
func RegisterPolicy(kind string, schema *PolicySchema) {
	policies[kind] = schema
}

//RegisterValidateFunc register a validate func for one type of value
//this API is not thread safe, only use it during sc init
func RegisterValidateFunc(t ValueType, f validate) {
	validators[t] = f
}
