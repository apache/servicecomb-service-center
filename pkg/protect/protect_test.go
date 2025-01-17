package protect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsWithinRestartProtection(t *testing.T) {
	restartProtectInterval = 2 * time.Minute

	// protection switch off
	enableInstanceNullProtect = false
	assert.False(t, ShouldProtectOnNullInstance())

	enableInstanceNullProtect = true
	AlwaysProtection()
	assert.True(t, ShouldProtectOnNullInstance())

	// protection delay exceed
	restartProtectInterval = 1 * time.Second
	StartProtectionAndStopDelayed()
	assert.True(t, ShouldProtectOnNullInstance())
	time.Sleep(restartProtectInterval)
	assert.False(t, ShouldProtectOnNullInstance())

	// only the first takes effects
	StartProtectionAndStopDelayed()
	assert.False(t, ShouldProtectOnNullInstance())
}
