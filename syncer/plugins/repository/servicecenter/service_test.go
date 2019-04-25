package servicecenter

import (
	"context"
	"testing"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
)

func TestClient_CreateService(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	_, err := repo.CreateService(context.Background(), "default/deault", &scpb.MicroService{})
	if err != nil {
		t.Errorf("create service failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.CreateService(context.Background(), "default/deault", &scpb.MicroService{})
	if err != nil {
		t.Logf("create service failed, error: %s", err)
	}
}

func TestClient_ServiceExistence(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	_, err := repo.ServiceExistence(context.Background(), "default/deault", &scpb.MicroService{})
	if err != nil {
		t.Errorf("check service existence failed, error: %s", err)
	}

	svr.Close()
	_, err = repo.ServiceExistence(context.Background(), "default/deault", &scpb.MicroService{})
	if err != nil {
		t.Logf("check service existence failed, error: %s", err)
	}
}

func TestClient_DeleteService(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	err := repo.DeleteService(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd")
	if err != nil {
		t.Errorf("delete service failed, error: %s", err)
	}

	svr.Close()
	err = repo.DeleteService(context.Background(), "default/deault",
		"4042a6a3e5a2893698ae363ea99a69eb63fc51cd")
	if err != nil {
		t.Logf("delete service failed, error: %s", err)
	}
}
