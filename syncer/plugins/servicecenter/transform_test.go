package servicecenter

import (
	pbsc "github.com/apache/servicecomb-service-center/syncer/proto/sc"
	scpb "github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestTransform_ServiceCopy(t *testing.T) {
	t.Run("value will copy from scpb.MicroService to pbsc.MicroService", func(t *testing.T) {
		service := scpb.MicroService{
			ServiceId: "1234567",
		}
		serviceInpbsc := ServiceCopy(&service)
		assert.NotNil(t, serviceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", serviceInpbsc.ServiceId)
	})
	t.Run("more values will copy from scpb.MicroService to pbsc.MicroService", func(t *testing.T) {
		service := scpb.MicroService{
			ServiceId:   "1234567",
			AppId:       "appid",
			ServiceName: "service",
		}
		serviceInpbsc := ServiceCopy(&service)
		assert.NotNil(t, serviceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", serviceInpbsc.ServiceId)
		assert.Equal(t, "appid", serviceInpbsc.AppId)
		assert.Equal(t, "service", serviceInpbsc.ServiceName)
	})
}

func TestTransform_ServiceCopyRe(t *testing.T) {
	t.Run("value will copy from pbsc.MicroService to scpb.MicroService", func(t *testing.T) {
		service := pbsc.MicroService{
			ServiceId: "1234567",
		}
		serviceInpbsc := ServiceCopyRe(&service)
		assert.NotNil(t, serviceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", serviceInpbsc.ServiceId)
	})
	t.Run("more values will copy from pbsc.MicroService to scpb.MicroService", func(t *testing.T) {
		service := pbsc.MicroService{
			ServiceId:   "1234567",
			AppId:       "appid",
			ServiceName: "service",
		}
		serviceInpbsc := ServiceCopyRe(&service)
		assert.NotNil(t, serviceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", serviceInpbsc.ServiceId)
		assert.Equal(t, "appid", serviceInpbsc.AppId)
		assert.Equal(t, "service", serviceInpbsc.ServiceName)
	})
}

func TestTransform_InstanceCopy(t *testing.T) {
	t.Run("value will copy from scpb.MicroServiceInstance to pbsc.MicroServiceInstance", func(t *testing.T) {
		instance := scpb.MicroServiceInstance{
			ServiceId: "1234567",
		}
		instanceInpbsc := InstanceCopy(&instance)
		assert.NotNil(t, instanceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", instanceInpbsc.ServiceId)
	})
	t.Run("more values will copy from scpb.MicroServiceInstance to pbsc.MicroServiceInstance", func(t *testing.T) {
		instance := scpb.MicroServiceInstance{
			ServiceId:  "1234567",
			InstanceId: "7654321",
		}
		instanceInpbsc := InstanceCopy(&instance)
		assert.NotNil(t, instanceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", instanceInpbsc.ServiceId)
		assert.Equal(t, "7654321", instanceInpbsc.InstanceId)
	})
}

func TestTransform_InstanceCopyRe(t *testing.T) {
	t.Run("value will copy from pbsc.MicroServiceInstance to scpb.MicroServiceInstance", func(t *testing.T) {
		instance := pbsc.MicroServiceInstance{
			ServiceId: "1234567",
		}
		instanceInpbsc := InstanceCopyRe(&instance)
		assert.NotNil(t, instanceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", instanceInpbsc.ServiceId)
	})
	t.Run("value will copy from pbsc.MicroServiceInstance to scpb.MicroServiceInstance", func(t *testing.T) {
		instance := pbsc.MicroServiceInstance{
			ServiceId:  "1234567",
			InstanceId: "7654321",
		}
		instanceInpbsc := InstanceCopyRe(&instance)
		assert.NotNil(t, instanceInpbsc.ServiceId, "serviceId is not nil")
		assert.Equal(t, "1234567", instanceInpbsc.ServiceId)
		assert.Equal(t, "7654321", instanceInpbsc.InstanceId)
	})
}

func TestTransform_SchemaCopy(t *testing.T) {
	t.Run("value will copy from scpb.Schema to pbsc.Schema", func(t *testing.T) {
		schema := scpb.Schema{
			SchemaId: "1234567",
		}
		instanceInpbsc := SchemaCopy(&schema)
		assert.NotNil(t, instanceInpbsc.SchemaId, "schemaId is not nil")
		assert.Equal(t, "1234567", instanceInpbsc.SchemaId)
	})
}
