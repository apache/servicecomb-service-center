package sd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeStore_Initialize(t *testing.T) {
	s := TypeStore{}
	s.Initialize()
	assert.NotNil(t, s.ready)
	assert.NotNil(t, s.goroutine)
	assert.Equal(t, false, s.isClose)
}

func TestTypeStore_Ready(t *testing.T) {
	s := TypeStore{}
	s.Initialize()
	c := s.Ready()
	assert.NotNil(t, c)
}

func TestTypeStore_Stop(t *testing.T) {
	t.Run("when closed", func(t *testing.T) {
		s := TypeStore{
			isClose: true,
		}
		s.Initialize()
		s.Stop()
		assert.Equal(t, true, s.isClose)
	})

	t.Run("when not closed", func(t *testing.T) {
		s := TypeStore{
			isClose: false,
		}
		s.Initialize()
		s.Stop()
		assert.Equal(t, true, s.isClose)
	})
}
