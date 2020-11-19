package etcd_test

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	a1 = rbacframe.Account{
		ID:                  "11111-22222-33333",
		Name:                "test-account1",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
	a2 = rbacframe.Account{
		ID:                  "11111-22222-33333-44444",
		Name:                "test-account2",
		Password:            "tnuocca-tset",
		Role:                "admin",
		TokenExpirationTime: "2020-12-30",
		CurrentPassword:     "tnuocca-tset1",
	}
)

func TestAccount(t *testing.T) {
	t.Run("add and get account", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "test-account-key", &a1)
		assert.NoError(t, err)
		r, err := datasource.Instance().GetAccount(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.Equal(t, a1, *r)
	})
	t.Run("account should exist", func(t *testing.T) {
		exist, err := datasource.Instance().AccountExist(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.True(t, exist)
	})
	t.Run("delete account", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "test-account-key222", &a1)
		assert.NoError(t, err)
		_, err = datasource.Instance().DeleteAccount(context.Background(), "test-account-key222")
		assert.NoError(t, err)
	})
	t.Run("add two accounts and list", func(t *testing.T) {
		err := datasource.Instance().UpdateAccount(context.Background(), "key1", &a1)
		assert.NoError(t, err)
		err = datasource.Instance().UpdateAccount(context.Background(), "key2", &a2)
		assert.NoError(t, err)
		accs, n, err := datasource.Instance().ListAccount(context.Background(), "key")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)
		t.Log(accs)
	})
}
