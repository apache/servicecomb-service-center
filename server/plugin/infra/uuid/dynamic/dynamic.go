package dynamic

import (
	"github.com/ServiceComb/service-center/pkg/plugin"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/pkg/uuid"
	mgr "github.com/ServiceComb/service-center/server/plugin"
)

var (
	serviceIdFunc  func() string
	instanceIdFunc func() string
)

func init() {
	serviceIdFunc = findUuidFunc("GetServiceId")
	instanceIdFunc = findUuidFunc("GetInstanceId")

	mgr.RegisterPlugin(mgr.Plugin{mgr.DYNAMIC, mgr.UUID, "dynamic", New})
}

func buildinUuidFunc() string {
	return uuid.GenerateUuid()
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

func New() mgr.PluginInstance {
	return &DynamicUUID{}
}

type DynamicUUID struct {
}

func (du *DynamicUUID) GetServiceId() string {
	return serviceIdFunc()
}

func (du *DynamicUUID) GetInstanceId() string {
	return instanceIdFunc()
}
