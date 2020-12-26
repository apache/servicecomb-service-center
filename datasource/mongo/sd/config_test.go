package sd

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	cfg := Config{
		Key: "",
	}
	assert.Empty(t, cfg, "config is empty")

	cfg1 := cfg.WithTable("configKey")
	assert.Equal(t, "configKey", cfg1.Key,
		"contain key after method WithTable")

	assert.Equal(t, 0, cfg1.InitSize,
		"init size is zero")
	cfg1 = cfg.WithInitSize(2)
	assert.Equal(t, 2, cfg1.InitSize,
		"init size is not zero after method WithInitSize")

	assert.Equal(t, time.Duration(0), cfg1.Timeout,
		"timeOut is zero")
	cfg1 = cfg.WithTimeout(6)
	assert.Equal(t, time.Duration(6), cfg1.Timeout,
		"timeOut is not zero after method WithTimeout")

	assert.Equal(t, time.Duration(0), cfg1.Period,
		"timeOut is zero")
	cfg1 = cfg.WithPeriod(8)
	assert.Equal(t, time.Duration(8), cfg1.Period,
		"timeOut is not zero after method WithTimeout")

	assert.Nil(t, cfg1.OnEvent,
		"EventFunc is nil before method EventFunc")
	mongoEventFunc = mongoEventFuncGet()
	cfg1 = cfg.WithEventFunc(mongoEventFunc)
	assert.NotNil(t, cfg1.OnEvent)

	cfg1 = cfg.AppendEventFunc(mongoEventFuncGetOther())
	assert.NotNil(t, cfg1.OnEvent)

	out := cfg1.String()
	assert.NotNil(t, out,
		"method String return not after methods")
}

var mongoEventFunc MongoEventFunc

func mongoEventFuncGet() MongoEventFunc{
	fun := func(evt MongoEvent) {
		evt.DocumentId = "DocumentId has changed"
		evt.BusinessId = "BusinessId has changed"
		evt.Value = 2
		evt.Type = discovery.EVT_UPDATE
		log.Info("in event func")
	}
	return fun
}

func mongoEventFuncGetOther() MongoEventFunc {
	fun := func(evt MongoEvent) {
		fmt.Print("math2")
	}
	return fun
}
