package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/server/health"
)

func TestReadiness(t *testing.T) {
	health.SetGlobalReadinessChecker(&health.NullChecker{})
	assert.NoError(t, Readiness(nil))
}
