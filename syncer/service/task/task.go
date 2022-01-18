package task

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	serverconfig "github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	dbconfig "github.com/go-chassis/cari/db/config"
	carisync "github.com/go-chassis/cari/sync"
)

func initDatabase() {
	kind := serverconfig.GetString("registry.kind", "", serverconfig.WithStandby("registry_plugin"))
	endpoint := serverconfig.GetString("registry.etcd.cluster.endpoints", "")
	log.Info(fmt.Sprintf("db endpoint is %s", endpoint))

	_, err := getDatasourceTLSConfig()
	if err != nil {
		log.Error("init database failed", err)
		return
	}

	if err := datasource.Init(&dbconfig.Config{
		Kind:        kind,
		URI:         endpoint,
		PoolSize:    5,
		SSLEnabled:  serverconfig.GetSSL().SslEnabled,
		VerifyPeer:  false,
		RootCA:      "",
		CertFile:    "",
		KeyFile:     "",
		CertPwdFile: "",
		Timeout:     0,
	}); err != nil {
		log.Fatal("init datasource failed", err)
	}
}

func getDatasourceTLSConfig() (*tls.Config, error) {
	if serverconfig.GetSSL().SslEnabled {
		return tlsconf.ClientConfig()
	}
	return nil, nil
}

func ListTask(ctx context.Context) ([]*carisync.Task, error) {
	return datasource.GetTaskDao().List(ctx)
}

type syncTasks []*carisync.Task

func (s syncTasks) Len() int {
	return len(s)
}

func (s syncTasks) Less(i, j int) bool {
	return s[i].Timestamp <= s[j].Timestamp
}

func (s syncTasks) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
