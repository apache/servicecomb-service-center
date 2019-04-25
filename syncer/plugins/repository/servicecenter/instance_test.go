package servicecenter

import (
	"context"
	"testing"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

func TestClient_RegisterInstance(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	_, err := repo.RegisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", &scpb.MicroServiceInstance{})
	if err != nil {
		t.Errorf("register instance failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.RegisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", &scpb.MicroServiceInstance{})
	if err != nil {
		t.Logf("register instance failed, error: %s", err)
	}
}

func TestClient_UnregisterInstance(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	err := repo.UnregisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Errorf("unregister instance failed, error: %s", err)
	}

	svr.Close()
	err = repo.UnregisterInstance(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Logf("unregister instance failed, error: %s", err)
	}
}

func TestClient_DiscoveryInstances(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	_, err := repo.DiscoveryInstances(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "default","testservice", "1.0.1")
	if err != nil {
		t.Errorf("discovery instances failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.DiscoveryInstances(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "default","testservice", "1.0.1")
	if err != nil {
		t.Logf("discovery instances failed, error: %s", err)
	}
}

func TestClient_Heartbeat(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	err := repo.Heartbeat(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Errorf("send instance heartbeat failed, error: %s", err)
	}

	svr.Close()
	err = repo.Heartbeat(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd", "7a6be9f861a811e9b3f6fa163eca30e0")
	if err != nil {
		t.Logf("send instance heartbeat, error: %s", err)
	}
}
