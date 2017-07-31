package store

import (
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
)

type Indexer interface {
	Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error)
}

type KvCacheIndexer struct {
	cache Cache
}

func (i *KvCacheIndexer) Search(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	data := i.cache.Data(registry.BytesToStringWithNoCopy(op.Key))
	kv := data.(*mvccpb.KeyValue)
	return &registry.PluginResponse{
		Action:    op.Action,
		Kvs:       []*mvccpb.KeyValue{kv},
		Count:     1,
		Revision:  i.cache.Version(),
		Succeeded: true,
	}, nil
}

func NewKvCacheIndexer(c Cache) *KvCacheIndexer {
	return &KvCacheIndexer{
		cache: c,
	}
}
