package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthChecker_AddResult(t *testing.T) {
	h := &HealthChecker{
		checkIntervalBySecond: 1,
		failureWindow:         NewHealthCheckWindow(8, 5),
		recoveryWindow:        NewHealthCheckWindow(4, 2),
		shouldTrustPeerServer: true,
	}
	// 全部true
	for i := 0; i < 10; i++ {
		h.AddResult(true)
		assert.True(t, h.failureWindow.IsHealthy())
		assert.True(t, h.recoveryWindow.IsHealthy())
		assert.True(t, h.ShouldTrustPeerServer())
	}

	// t t t t f f f f
	h.AddResult(false)
	h.AddResult(false)
	h.AddResult(false)
	h.AddResult(false)
	assert.True(t, h.failureWindow.IsHealthy())
	assert.False(t, h.recoveryWindow.IsHealthy()) // recoveryWindow 首先变成失败
	assert.True(t, h.ShouldTrustPeerServer())     // 结果还是健康，因为 健康 > 不健康，看 failureWindow

	// t t t f f f f f
	h.AddResult(false)
	assert.False(t, h.failureWindow.IsHealthy())
	assert.False(t, h.recoveryWindow.IsHealthy())
	assert.False(t, h.ShouldTrustPeerServer()) // 不健康

	// 全false
	for i := 0; i < 10; i++ {
		h.AddResult(false)
		assert.False(t, h.failureWindow.IsHealthy())
		assert.False(t, h.recoveryWindow.IsHealthy())
		assert.False(t, h.ShouldTrustPeerServer())
	}
	assert.ElementsMatch(t, []bool{false, false, false, false, false, false, false, false}, h.failureWindow.checkPassResults)
	assert.ElementsMatch(t, []bool{false, false, false, false}, h.recoveryWindow.checkPassResults)

	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	assert.ElementsMatch(t, []bool{true, false, true, false, true, false, true, false}, h.failureWindow.checkPassResults)
	assert.ElementsMatch(t, []bool{true, false, true, false}, h.recoveryWindow.checkPassResults)
	assert.True(t, h.failureWindow.IsHealthy())   // failureWindow 恢复
	assert.False(t, h.recoveryWindow.IsHealthy()) // 但是 recoveryWindow 还没恢复
	assert.False(t, h.ShouldTrustPeerServer())    // 结果是不健康，因为不健康 > 健康，看 recoveryWindow

	h.AddResult(true)
	h.AddResult(true)
	assert.True(t, h.failureWindow.IsHealthy())
	assert.True(t, h.recoveryWindow.IsHealthy())
	assert.True(t, h.ShouldTrustPeerServer()) // 结果健康
}
