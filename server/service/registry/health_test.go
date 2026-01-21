package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/server/health"
	"github.com/apache/servicecomb-service-center/syncer/config"
)

func TestReadiness(t *testing.T) {
	config.SetConfig(config.Config{
		Sync: &config.Sync{},
	})
	health.SetGlobalReadinessChecker(&health.NullChecker{})
	assert.NoError(t, Readiness(nil))
}
