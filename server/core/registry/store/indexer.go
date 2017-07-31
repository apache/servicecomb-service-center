package store

import (
	"github.com/ServiceComb/service-center/server/core/registry"
	"golang.org/x/net/context"
)

type Indexer interface {
	Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error)
}

type KvCacheIndexer struct {
}

func (i *KvCacheIndexer) Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	return nil, nil
}

func NewKvCacheIndexer(c Cache) *KvCacheIndexer {
	return &KvCacheIndexer{}
}
