package kvstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCacheDomainProjectKey(t *testing.T) {
	testCache := new(KvCache)
	str := "/cse-sr/inst/files/default/default/heheheh/xixixi/"
	res := testCache.getCacheDomainProjectKey(str)
	assert.Equal(t, "/cse-sr/inst/files/default/default/", res)
}
