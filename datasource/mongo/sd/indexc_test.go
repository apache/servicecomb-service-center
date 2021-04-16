package sd_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/stretchr/testify/assert"
)

func TestIndexCache(t *testing.T) {
	indexCache := sd.NewIndexCache()
	assert.NotNil(t, indexCache)
	indexCache.Put("index1", "doc1")
	assert.Equal(t, []string{"doc1"}, indexCache.Get("index1"))
	indexCache.Put("index1", "doc2")
	assert.Len(t, indexCache.Get("index1"), 2)
	indexCache.Delete("index1", "doc1")
	assert.Len(t, indexCache.Get("index1"), 1)
	indexCache.Delete("index1", "doc2")
	assert.Nil(t, indexCache.Get("index1"))
}
