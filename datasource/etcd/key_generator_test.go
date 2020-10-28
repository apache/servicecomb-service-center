package etcd_test

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateETCDAccountKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/accounts/admin", etcd.GenerateETCDAccountKey("admin"))
}

func TestGenerateETCDProjectKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/projects/domain/project", etcd.GenerateETCDProjectKey("domain", "project"))
}

func TestGenerateETCDDomainKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/domains/domain", etcd.GenerateETCDDomainKey("domain"))
}
