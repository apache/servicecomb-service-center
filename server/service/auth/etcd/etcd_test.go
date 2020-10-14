package etcd

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/apache/servicecomb-service-center/server/service/auth"
	"github.com/astaxie/beego"
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

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.REGISTRY, Name: "etcd", New: etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.DISCOVERY, Name: "buildin", New: etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.DISCOVERY, Name: "etcd", New: etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.TRACING, Name: "buildin", New: pzipkin.New})
	auth.Install("etcd", func(opts auth.Options) (auth.DataSource, error) {
		return NewDataSource(), nil
	})
	err := auth.Init(auth.Options{
		Endpoint:       "",
		PluginImplName: "etcd",
	})
	if err != nil {
		panic("failed to register etcd auth plugin")
	}
}

func TestAccount(t *testing.T) {
	t.Run("add and get account", func(t *testing.T) {
		err := auth.Auth().UpdateAccount(context.Background(), "test-account-key", &a1)
		assert.NoError(t, err)
		r, err := auth.Auth().GetAccount(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.Equal(t, a1, *r)
	})
	t.Run("account should exist", func(t *testing.T) {
		exist, err := auth.Auth().AccountExist(context.Background(), "test-account-key")
		assert.NoError(t, err)
		assert.True(t, exist)
	})
	t.Run("delete account", func(t *testing.T) {
		err := auth.Auth().UpdateAccount(context.Background(), "test-account-key222", &a1)
		assert.NoError(t, err)
		_, err = auth.Auth().DeleteAccount(context.Background(), "test-account-key222")
		assert.NoError(t, err)
	})
	t.Run("add two accounts and list", func(t *testing.T) {
		err := auth.Auth().UpdateAccount(context.Background(), "key1", &a1)
		assert.NoError(t, err)
		err = auth.Auth().UpdateAccount(context.Background(), "key2", &a2)
		assert.NoError(t, err)
		accs, n, err := auth.Auth().ListAccount(context.Background(), "key")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), n)
		t.Log(accs)
	})
}

func TestDomain(t *testing.T) {
	t.Run("test domain", func(t *testing.T) {
		_, err := auth.Auth().AddDomain(context.Background(), "test-domain")
		assert.NoError(t, err)
		r, err := auth.Auth().DomainExist(context.Background(), "test-domain")
		assert.NoError(t, err)
		assert.Equal(t, true, r)
	})
}

func TestProject(t *testing.T) {
	t.Run("test project", func(t *testing.T) {
		_, err := auth.Auth().AddProject(context.Background(), "test-domain", "test-project")
		assert.NoError(t, err)
		r, err := auth.Auth().ProjectExist(context.Background(), "test-domain", "test-project")
		assert.NoError(t, err)
		assert.Equal(t, true, r)
	})
}
