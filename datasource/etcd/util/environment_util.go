package util

import (
	"context"

	ev "github.com/go-chassis/cari/env"
	"github.com/go-chassis/etcdadpt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/etcd/sd"
	"github.com/apache/servicecomb-service-center/datasource/etcd/state/kvstore"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func GetAllEnvironmentUtil(ctx context.Context) ([]*ev.Environment, error) {
	domainProject := util.ParseDomainProject(ctx)
	envs, err := GetEnvironmentsByDomainProject(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	return envs, nil
}

func GetEnvironmentsByDomainProject(ctx context.Context, domainProject string) ([]*ev.Environment, error) {
	kvs, err := getEnvironmentsRawData(ctx, domainProject)
	if err != nil {
		return nil, err
	}
	envs := []*ev.Environment{}
	for _, kv := range kvs {
		envs = append(envs, kv.Value.(*ev.Environment))
	}
	return envs, nil
}

func getEnvironmentsRawData(ctx context.Context, domainProject string) ([]*kvstore.KeyValue, error) {
	key := path.GenerateEnvironmentKey(domainProject, "")
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithPrefix())
	resp, err := sd.Environment().Search(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return resp.Kvs, err
}

func GetOneDomainProjectEnvironmentCount(ctx context.Context, domainProject string) (int64, error) {
	key := path.GetEnvironmentRootKey(domainProject) + path.SPLIT
	opts := append(FromContext(ctx),
		etcdadpt.WithStrKey(key),
		etcdadpt.WithCountOnly(),
		etcdadpt.WithPrefix())
	resp, err := sd.Environment().Search(ctx, opts...)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
