package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckWindow_AddResult(t *testing.T) {
	c := NewHealthCheckWindow(5, 2)
	for i := 0; i < 5; i++ {
		c.AddResult(true)
		assert.Equal(t, i+1, len(c.checkPassResults))
	}
	for i := 0; i < 5; i++ {
		c.AddResult(true)
		assert.Equal(t, c.windowSize, len(c.checkPassResults))
	}
	c.AddResult(false)
	c.AddResult(false)
	assert.ElementsMatch(t, []bool{true, true, true, false, false}, c.checkPassResults)
}

func TestHealthCheckWindow_IsHealthy(t *testing.T) {
	c := NewHealthCheckWindow(4, 2)

	// f 健康
	c.AddResult(false)
	assert.True(t, c.IsHealthy())

	// 达到2次失败，不健康
	for i := 0; i < 5; i++ {
		c.AddResult(false)
		assert.False(t, c.IsHealthy())
	}
	// 结束状态：f f f f

	// f f f t 3次失败
	c.AddResult(true)
	assert.False(t, c.IsHealthy())

	// f f t t 2次失败
	c.AddResult(true)
	assert.False(t, c.IsHealthy())

	// f t t t 1次失败，变为健康
	c.AddResult(true)
	assert.True(t, c.IsHealthy())

	// t t t f 1次失败
	c.AddResult(false)
	assert.True(t, c.IsHealthy())

	// t t f f 2次失败，变为不健康
	c.AddResult(false)
	assert.False(t, c.IsHealthy())

	// t f f t 2次失败
	c.AddResult(true)
	assert.False(t, c.IsHealthy())

	// f f t t 2次失败
	c.AddResult(true)
	assert.False(t, c.IsHealthy())

	// 小于2次失败，变为成功
	for i := 0; i < 4; i++ {
		c.AddResult(true)
		assert.True(t, c.IsHealthy())
	}
}
