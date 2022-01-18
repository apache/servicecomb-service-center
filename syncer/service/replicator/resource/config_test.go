package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	c := NewConfig(nil)
	assert.NotNil(t, c)
	ctx := context.TODO()
	t.Run("LoadCurrentResource", func(t *testing.T) {
		result := c.LoadCurrentResource(ctx)
		assert.Equal(t, result, NonImplementResult())
	})
	t.Run("NeedOperate", func(t *testing.T) {
		result := c.NeedOperate(ctx)
		assert.Equal(t, result, NonImplementResult())
	})
	t.Run("Operate", func(t *testing.T) {
		result := c.Operate(ctx)
		assert.Equal(t, result, NonImplementResult())
	})
}
