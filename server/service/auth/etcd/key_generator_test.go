package etcd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateETCDAccountKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/accounts/admin", GenerateETCDAccountKey("admin"))
}

func TestGenerateETCDProjectKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/projects/domain/project", GenerateETCDProjectKey("domain", "project"))
}

func TestGenerateETCDDomainKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/domains/domain", GenerateETCDDomainKey("domain"))
}
