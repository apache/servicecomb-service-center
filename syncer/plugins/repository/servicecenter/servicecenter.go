package servicecenter

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const PluginName = "servicecenter"

func init() {
	plugins.RegisterPlugin(&plugins.Plugin{
		Kind: plugins.PluginRepository,
		Name: PluginName,
		New:  New,
	})
}

type adaptor struct{}

func New() plugins.PluginInstance {
	return &adaptor{}
}

func (*adaptor) New(endpoints []string) (repository.Repository, error) {
	cli, err := sc.NewSCClient(sc.Config{Endpoints: endpoints})
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

type Client struct {
	cli *sc.SCClient
}

func (c *Client) GetAll(ctx context.Context) (*pb.SyncData, error) {
	cache, err := c.cli.GetScCache(ctx)
	if err != nil {
		return nil, err
	}
	return transform(cache), nil

}
