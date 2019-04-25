package peer

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/serf/serf"
)

func TestAgent(t *testing.T) {
	conf := DefaultConfig()
	agent, err := Create(conf, nil)
	if err != nil {
		t.Errorf("create agent failed, error: %s", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	agent.Start(ctx)

	go func() {
		agent.ShutdownCh()
	}()
	time.Sleep(time.Second)

	err = agent.UserEvent("test", []byte("test"), true)
	if err != nil {
		t.Errorf("send user event failed, error: %s", err)
	}

	_, err = agent.Query("test", []byte("test"), &serf.QueryParam{})
	if err != nil {
		t.Errorf("query for other node failed, error: %s", err)
	}
	agent.LocalMember()

	agent.Member("testnode")

	agent.SerfConfig()

	_, err = agent.Join([]string{"127.0.0.1:9999"}, true)
	if err != nil {
		t.Logf("join to other node failed, error: %s", err)
	}

	err = agent.Leave()
	if err != nil {
		t.Errorf("angent leave failed, error: %s", err)
	}

	err = agent.ForceLeave("testnode")
	if err != nil {
		t.Errorf("angent force leave failed, error: %s", err)
	}

	err = agent.Shutdown()
	if err != nil {
		t.Errorf("angent shutdown failed, error: %s", err)
	}
	cancel()
}
