package resource

import (
	"context"
	"errors"
	"sync"

	svcconfig "github.com/apache/servicecomb-service-center/server/config"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/little-cui/etcdadpt"
)

const (
	KV = "kv"
)

const (
	KVKey         = "key"
	KVKeyNonExist = "key not exist in opts"
)

var (
	manager KeyManager

	ErrNotImplement   = errors.New("not implement")
	ErrRecordNonExist = errors.New("record non exist")
)

func NewKV(e *v1sync.Event) Resource {
	r := &kv{
		event:   e,
		manager: keyManage(),
	}
	return r
}

type kv struct {
	event *v1sync.Event
	key   string

	manager         KeyManager
	tombstoneLoader tombstoneLoader

	cur []byte

	defaultFailHandler
}

func (k *kv) LoadCurrentResource(ctx context.Context) *Result {
	key, ok := k.event.Opts[KVKey]
	if !ok {
		return NewResult(Fail, KVKeyNonExist)
	}
	k.key = key

	value, err := k.manager.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrRecordNonExist) {
			return nil
		}
		return FailResult(err)
	}
	k.cur = value
	return nil
}

func (k *kv) NeedOperate(ctx context.Context) *Result {
	c := &checker{
		curNotNil:  k.cur != nil,
		event:      k.event,
		updateTime: nil,
		resourceID: k.key,
	}
	c.tombstoneLoader = c
	if k.tombstoneLoader != nil {
		c.tombstoneLoader = k.tombstoneLoader
	}

	return c.needOperate(ctx)
}

func (k *kv) CreateHandle(ctx context.Context) error {
	return k.manager.Post(ctx, k.key, k.event.Value)
}

func (k *kv) UpdateHandle(ctx context.Context) error {
	return k.manager.Put(ctx, k.key, k.event.Value)
}

func (k *kv) DeleteHandle(ctx context.Context) error {
	return k.manager.Delete(ctx, k.key)
}

var once sync.Once

func keyManage() KeyManager {
	once.Do(InitManager)
	return manager
}

func (k *kv) Operate(ctx context.Context) *Result {
	return newOperator(k).operate(ctx, k.event.Action)
}

type KeyManager interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Post(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}

type mongoManager struct {
}

func (m *mongoManager) Get(_ context.Context, _ string) ([]byte, error) {
	return nil, ErrNotImplement
}

func (m *mongoManager) Put(_ context.Context, _ string, _ []byte) error {
	return ErrNotImplement
}

func (m *mongoManager) Post(_ context.Context, _ string, _ []byte) error {
	return ErrNotImplement
}

func (m *mongoManager) Delete(_ context.Context, _ string) error {
	return ErrNotImplement
}

type etcdManager struct {
}

func InitManager() {
	kind := svcconfig.GetString("registry.kind", "",
		svcconfig.WithStandby("registry_plugin"))
	if kind == "etcd" {
		manager = new(etcdManager)
		return
	}

	manager = new(mongoManager)
}

func (e *etcdManager) Get(ctx context.Context, key string) ([]byte, error) {
	r, err := etcdadpt.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, ErrRecordNonExist
	}

	return r.Value, nil
}

func (e etcdManager) Put(ctx context.Context, key string, value []byte) error {
	return etcdadpt.Put(ctx, key, string(value))
}

func (e etcdManager) Post(ctx context.Context, key string, value []byte) error {
	return etcdadpt.Put(ctx, key, string(value))
}

func (e etcdManager) Delete(ctx context.Context, key string) error {
	_, err := etcdadpt.Delete(ctx, key)
	return err
}
