package dynamic

import (
	"github.com/ServiceComb/service-center/pkg/plugin"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/pkg/uuid"
	"strings"
)

func buildinUuidFunc() string {
	return strings.Replace(uuid.NewV1().String(), string(uuid.DASH), "", -1)
}

func findUuidFunc(funcName string) func() string {
	ff, err := plugin.FindFunc("uuid", funcName)
	if err != nil {
		return buildinUuidFunc
	}
	f, ok := ff.(func() string)
	if !ok {
		util.Logger().Warnf(nil, "unexpected function '%s' format found in plugin 'uuid'.", funcName)
		return buildinUuidFunc
	}
	return f
}

func GetServiceId() string {
	f := findUuidFunc("GetServiceId")
	return f()
}

func GetInstanceId() string {
	f := findUuidFunc("GetInstanceId")
	return f()
}

func GenerateUuid() string {
	return buildinUuidFunc()
}
