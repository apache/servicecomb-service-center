//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package registry

import (
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

var (
	RegistryPlugins  map[string]func(cfg *Config) Registry
	registryInstance Registry
	singletonLock    sync.Mutex
)

type ActionType int
type SortOrder int
type CompareType int
type CompareResult int

const (
	GET    ActionType = 0
	PUT    ActionType = 1
	DELETE ActionType = 2

	SORT_NONE    SortOrder = 0
	SORT_ASCEND  SortOrder = 1
	SORT_DESCEND SortOrder = 2

	CMP_VERSION CompareType = 0
	CMP_CREATE  CompareType = 1
	CMP_MOD     CompareType = 2
	CMP_VALUE   CompareType = 3

	CMP_EQUAL     CompareResult = 0
	CMP_GREATER   CompareResult = 1
	CMP_LESS      CompareResult = 2
	CMP_NOT_EQUAL CompareResult = 3

	REFRESH_MANAGER_CLUSTER_INTERVAL = 30

	REQUEST_TIMEOUT = 300
)

type Registry interface {
	Err() <-chan error
	Ready() <-chan int
	PutNoOverride(ctx context.Context, op *PluginOp) (bool, error)
	Do(ctx context.Context, op *PluginOp) (*PluginResponse, error)
	Txn(ctx context.Context, ops []*PluginOp) (*PluginResponse, error)
	TxnWithCmp(ctx context.Context, success []*PluginOp, cmp []*CompareOp, fail []*PluginOp) (*PluginResponse, error)
	LeaseGrant(ctx context.Context, TTL int64) (leaseID int64, err error)
	LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error)
	LeaseRevoke(ctx context.Context, leaseID int64) error
	// this function block util:
	// 1. connection error
	// 2. call send function failed
	// 3. response.Err()
	// 4. time out to watch, but return nil
	Watch(ctx context.Context, op *PluginOp, send func(message string, evt *PluginResponse) error) error
	Close()
	CompactCluster(ctx context.Context)
	Compact(ctx context.Context, revision int64) error
}

type Store interface {
}

type Config struct {
	EmbedMode        string
	ClusterAddresses string
	AutoSyncInterval int64
}

type PluginOp struct {
	Action     ActionType
	Key        []byte
	EndKey     []byte
	Value      []byte
	WithPrefix bool
	WithPrevKV bool
	Lease      int64
	KeyOnly    bool
	CountOnly  bool
	SortOrder  SortOrder
	WithRev    int64
	WithCache  bool
}

type PluginResponse struct {
	Action    ActionType
	Kvs       []*mvccpb.KeyValue
	PrevKv    *mvccpb.KeyValue
	Count     int64
	Revision  int64
	Succeeded bool
}

type CompareOp struct {
	Key    []byte
	Type   CompareType
	Result CompareResult
	Value  interface{}
}

func init() {
	RegistryPlugins = make(map[string]func(cfg *Config) Registry)
}

func GetRegisterCenter() Registry {
	if registryInstance == nil {
		singletonLock.Lock()
		if registryInstance == nil {
			registryFunc := RegistryPlugins[beego.AppConfig.String("registry_plugin")]
			autoSyncInterval, _ := beego.AppConfig.Int64("auto_sync_interval")
			if autoSyncInterval <= 0 {
				autoSyncInterval = REFRESH_MANAGER_CLUSTER_INTERVAL
			}
			registryInstance = registryFunc(&Config{
				ClusterAddresses: beego.AppConfig.String("manager_cluster"),
				AutoSyncInterval: autoSyncInterval,
			})
			select {
			case err := <-registryInstance.Err():
				panic(err)
			case <-registryInstance.Ready():
			}
		}
		singletonLock.Unlock()
	}
	return registryInstance
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, REQUEST_TIMEOUT*time.Second)
}

func WithWatchPrefix(key string) *PluginOp {
	return &PluginOp{
		Action:     GET,
		Key:        []byte(key),
		WithPrefix: true,
		WithPrevKV: true,
	}
}

func BytesToStringWithNoCopy(bytes []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}
