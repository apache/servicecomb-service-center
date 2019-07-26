package mocksotrage

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net/url"
	"os"

	"github.com/apache/servicecomb-service-center/syncer/etcd"
)

const (
	defaultName           = "etcd_mock"
	defaultDataDir        = "mock-data/"
	defaultListenPeerAddr = "http://127.0.0.1:30993"
)

type MockServer struct {
	etcd *etcd.Agent
}

func NewKVServer() (svr *MockServer, err error) {
	agent := etcd.NewAgent(defaultConfig())
	go agent.Start(context.Background())
	select {
	case <-agent.Ready():
	case err = <-agent.Error():
	}
	if err != nil {
		return nil, err
	}
	return &MockServer{agent}, nil
}

func (m *MockServer) Storage() *clientv3.Client {
	return m.etcd.Storage()
}

func (m *MockServer)Stop()  {
	m.etcd.Stop()
	os.RemoveAll(defaultDataDir)
}

func defaultConfig() *etcd.Config {
	peer, _ := url.Parse(defaultListenPeerAddr)
	conf := etcd.DefaultConfig()
	conf.Name = defaultName
	conf.Dir = defaultDataDir + defaultName
	conf.APUrls = []url.URL{*peer}
	conf.LPUrls = []url.URL{*peer}
	conf.InitialCluster = fmt.Sprintf("%s=%s", defaultName, defaultListenPeerAddr)
	return conf
}
