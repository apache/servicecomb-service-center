package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthChecker_AddResult(t *testing.T) {
	h := &HealthChecker{
		checkIntervalBySecond: 1,
		checkWindow:           NewHealthCheckWindow(8, 5),
		syncRecoveryWindow:    NewHealthCheckWindow(4, 2),
		shouldTrustPeerServer: true,
	}
	// 全部true
	for i := 0; i < 10; i++ {
		h.AddResult(true)
		assert.True(t, h.checkWindow.IsHealthy())
		assert.True(t, h.syncRecoveryWindow.IsHealthy())
		assert.True(t, h.ShouldTrustPeerServer())
	}

	// t t t t f f f f
	h.AddResult(false)
	h.AddResult(false)
	h.AddResult(false)
	h.AddResult(false)
	assert.True(t, h.checkWindow.IsHealthy())
	assert.False(t, h.syncRecoveryWindow.IsHealthy()) // sync recovery window首先变成失败
	assert.True(t, h.ShouldTrustPeerServer())         // 结果还是健康，因为 健康 > 不健康，看checkWindow

	// t t t f f f f f
	h.AddResult(false)
	assert.False(t, h.checkWindow.IsHealthy())
	assert.False(t, h.syncRecoveryWindow.IsHealthy())
	assert.False(t, h.ShouldTrustPeerServer()) // 不健康

	// 全false
	for i := 0; i < 10; i++ {
		h.AddResult(false)
		assert.False(t, h.checkWindow.IsHealthy())
		assert.False(t, h.syncRecoveryWindow.IsHealthy())
		assert.False(t, h.ShouldTrustPeerServer())
	}
	assert.ElementsMatch(t, []bool{false, false, false, false, false, false, false, false}, h.checkWindow.checkPassResults)
	assert.ElementsMatch(t, []bool{false, false, false, false}, h.syncRecoveryWindow.checkPassResults)

	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	h.AddResult(true)
	h.AddResult(false)
	assert.ElementsMatch(t, []bool{true, false, true, false, true, false, true, false}, h.checkWindow.checkPassResults)
	assert.ElementsMatch(t, []bool{true, false, true, false}, h.syncRecoveryWindow.checkPassResults)
	assert.True(t, h.checkWindow.IsHealthy())         // checkWindow恢复
	assert.False(t, h.syncRecoveryWindow.IsHealthy()) // 但是syncRecoveryWindow还没恢复
	assert.False(t, h.ShouldTrustPeerServer())        // 结果是不健康，因为不健康 > 健康，需要 checkWindow和 syncRecoveryWindow均为健康

	h.AddResult(true)
	h.AddResult(true)
	assert.True(t, h.checkWindow.IsHealthy())
	assert.True(t, h.syncRecoveryWindow.IsHealthy())
	assert.True(t, h.ShouldTrustPeerServer()) // 结果健康
}
