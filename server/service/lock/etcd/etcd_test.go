package etcd

import (
	"fmt"
	_ "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	_ "github.com/apache/servicecomb-service-center/server/plugin/tracing/pzipkin"
	"github.com/apache/servicecomb-service-center/server/service/lock"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	IsDebug = true
	lock.Install("etcd", func(opts lock.Options) (lock.DataSource, error) {
		return NewDataSource(), nil
	})
	err := lock.Init(lock.Options{
		Endpoint:       "",
		PluginImplName: "etcd",
	})
	if err != nil {
		panic("register etcd-lock plugin failed")
	}
}

func TestDLock(t *testing.T) {
	m1, err := lock.Lock().NewDLock("key1", 5, true)
	assert.NoError(t, err)
	assert.NotNil(t, m1)
	t.Log("m1 locked")

	ch := make(chan bool)
	go func() {
		m2, err := lock.Lock().NewDLock("key1", 1, false)
		assert.Nil(t, m2)
		assert.Error(t, err)
		fmt.Println("m2 try failed")
		m2, err = lock.Lock().NewDLock("key1", 1, true)
		assert.Nil(t, m2)
		assert.Error(t, err)
		fmt.Println("m2 timed out")
		ch <- true
	}()
	<-ch
	m3, err := lock.Lock().NewDLock("key1", 2, true)
	assert.NoError(t, err)
	assert.NotNil(t, m3)
	fmt.Println("m3 locked")
	err = m3.Unlock()
	assert.NoError(t, err)

	err = m1.Unlock()
	assert.NoError(t, err)
	fmt.Println("m1 unlocked")
}
