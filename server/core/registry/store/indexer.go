package store

import (
	"context"
	"github.com/ServiceComb/service-center/server/core/registry"
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
