package gov

import "fmt"

var validators = map[ValueType]validate{
	ValueTypeInt:    validateInt,
	ValueTypeString: validateString,
	ValueTypeList:   validateList,
	ValueTypeBool:   validateBool,
}

func validateAny(k string, v interface{}, attribute *PolicyAttribute) error {
	if attribute.Required && v == nil {
		return fmt.Errorf(MsgRequired, k)
	}
	f, ok := validators[attribute.Type]
	if !ok {
		return fmt.Errorf(MsgUnknownType, k)
	}
	return f(k, v, attribute)
}

func validateInt(k string, v interface{}, attribute *PolicyAttribute) error {
	integer, ok := v.(int)
	if !ok {
		fmt.Sprintf(MsgTypeInt, k)
	}
	return calIntRange(k, integer, attribute.Min, attribute.Max)
}
func validateBool(k string, v interface{}, attribute *PolicyAttribute) error {
	_, ok := v.(bool)
	if !ok {
		fmt.Sprintf(MsgTypeBool, k)
	}
	return nil
}
func validateString(k string, v interface{}, attribute *PolicyAttribute) error {
	s, ok := v.(string)
	if !ok {
		fmt.Sprintf(MsgTypeInt, k)
	}
	return calIntRange(k, len(s), attribute.Min, attribute.Max)
}
func validateList(k string, v interface{}, attribute *PolicyAttribute) error {
	switch t := v.(type) {
	case []string:
		return calIntRange(k, len(t), attribute.Min, attribute.Max)
	case []int:
		return calIntRange(k, len(t), attribute.Min, attribute.Max)
	default:
		return fmt.Errorf("can not support type %s", t)
	}

	return nil
}

func calIntRange(k string, value, min, max int) error {
	if min != 0 {
		if value < min {
			return fmt.Errorf(MsgTooSmall, k, min)
		}
	}
	if max != 0 {
		if value > max {
			return fmt.Errorf(MsgTooBig, k, max)
		}
	}
	return nil
}
