package sd

import (
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	options := Options{
		Key: "",
	}
	assert.Empty(t, options, "config is empty")

	options1 := options.SetTable("configKey")
	assert.Equal(t, "configKey", options1.Key,
		"contain key after method WithTable")

	assert.Equal(t, 0, options1.InitSize,
		"init size is zero")

	mongoEventFunc = mongoEventFuncGet()

	out := options1.String()
	assert.NotNil(t, out,
		"method String return not after methods")
}

var mongoEventFunc MongoEventFunc

func mongoEventFuncGet() MongoEventFunc {
	fun := func(evt MongoEvent) {
		evt.DocumentID = "DocumentID has changed"
		evt.ResourceID = "BusinessID has changed"
		evt.Value = 2
		evt.Type = discovery.EVT_UPDATE
		log.Info("in event func")
	}
	return fun
}
