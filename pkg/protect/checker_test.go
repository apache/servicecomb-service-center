package protect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDelayedStopProtectChecker_CheckProtection(t *testing.T) {
	p := NewDelayedSuccessChecker(1 * time.Second)
	assert.True(t, p.CheckProtection())
	time.Sleep(p.delay)
	assert.False(t, p.CheckProtection())
}
