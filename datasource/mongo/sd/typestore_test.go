package sd

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
	s.Run()
}

func TestTypeStore_getOrCreateCache(t *testing.T) {
	t.Run("get service cacher", func(t *testing.T) {
		s := TypeStore{}
		mc := s.getOrCreateCache("service")
		assert.NotNil(t, mc)
		assert.Equal(t, "service", mc.cache.name)
	})
	t.Run("get instance cacher", func(t *testing.T) {
		s := TypeStore{}
		mc := s.getOrCreateCache("instance")
		assert.NotNil(t, mc)
		assert.Equal(t, "instance", mc.cache.name)
	})
}

func TestTypeStore_autoClearCache(t *testing.T) {

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

func TestTypeStore_TypeCacher(t *testing.T) {
	t.Run("get service cacher", func(t *testing.T) {
		s := TypeStore{}
		mc := s.Service()
		assert.NotNil(t, mc)
		assert.Equal(t, "service", mc.cache.name)
	})
	t.Run("get instance cacher", func(t *testing.T) {
		s := TypeStore{}
		mc := s.Instance()
		assert.NotNil(t, mc)
		assert.Equal(t, "instance", mc.cache.name)
	})
}