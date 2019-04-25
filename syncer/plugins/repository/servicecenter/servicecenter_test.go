package servicecenter

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	"github.com/apache/servicecomb-service-center/syncer/test/datacenter/servicecenter"
)

func TestClient_GetAll(t *testing.T) {
	svr, repo := newServiceCenterRepo(t)
	_, err := repo.GetAll(context.Background())
	if err != nil {
		t.Errorf("get all from %s server failed, error: %s", PluginName, err)
	}

	svr.Close()
	_, err = repo.GetAll(context.Background())
	if err != nil {
		t.Logf("get all from %s server failed, error: %s", PluginName, err)
	}
}

func newServiceCenterRepo(t *testing.T) (*httptest.Server, repository.Repository) {
	plugins.SetPluginConfig(plugins.PluginRepository.String(), PluginName)
	adaptor := plugins.Plugins().Repository()
	if adaptor == nil {
		t.Errorf("get repository adaptor %s failed", PluginName)
	}
	svr := servicecenter.NewMockServer()
	if svr == nil {
		t.Error("new httptest server failed")
	}
	repo, err := adaptor.New([]string{svr.URL})
	if err != nil {
		t.Errorf("new repository %s failed, error: %s", PluginName, err)
	}
	return svr, repo
}
