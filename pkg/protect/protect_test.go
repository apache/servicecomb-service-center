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
	assert.False(t, IsWithinRestartProtection())
	// within protection
	enableInstanceNullProtect = true
	isWithinProtection = true
	startupTimestamp = time.Now().Add(-1 * time.Minute).UnixNano()
	assert.True(t, IsWithinRestartProtection())

	// protection delay exceed
	enableInstanceNullProtect = true
	isWithinProtection = true
	startupTimestamp = time.Now().Add(-2 * time.Minute).Unix()
	assert.False(t, IsWithinRestartProtection())

	// always false after exceed
	assert.False(t, IsWithinRestartProtection())
}
