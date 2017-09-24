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
	"encoding/json"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strconv"
	"sync"
	"time"
)

var (
	RegistryPlugins  map[string]func(cfg *Config) Registry
	registryInstance Registry
	singletonLock    sync.Mutex
	wait_delay       = []int{1, 1, 5, 10, 20, 30, 60}
)

type ActionType int

func (at ActionType) String() string {
	switch at {
	case Get:
		return "GET"
	case Put:
		return "PUT"
	case Delete:
		return "DELETE"
	default:
		return "ACTION" + strconv.Itoa(int(at))
	}
}

type CacheMode int

func (cm CacheMode) String() string {
	switch cm {
	case MODE_BOTH:
		return "MODE_BOTH"
	case MODE_CACHE:
		return "MODE_CACHE"
	case MODE_NO_CACHE:
		return "MODE_NO_CACHE"
	default:
		return "MODE" + strconv.Itoa(int(cm))
	}
}

type SortOrder int

func (so SortOrder) String() string {
	switch so {
	case SORT_NONE:
		return "SORT_NONE"
	case SORT_ASCEND:
		return "SORT_ASCEND"
	case SORT_DESCEND:
		return "SORT_DESCEND"
	default:
		return "SORT" + strconv.Itoa(int(so))
	}
}

type CompareType int
type CompareResult int

const (
	Get ActionType = iota
	Put
	Delete
)

const (
	SORT_NONE SortOrder = iota
	SORT_ASCEND
	SORT_DESCEND
)

const (
	CMP_VERSION CompareType = iota
	CMP_CREATE
	CMP_MOD
	CMP_VALUE
)

const (
	CMP_EQUAL CompareResult = iota
	CMP_GREATER
	CMP_LESS
	CMP_NOT_EQUAL
)

const (
	MODE_BOTH CacheMode = iota
	MODE_CACHE
	MODE_NO_CACHE
)

const (
	REFRESH_MANAGER_CLUSTER_INTERVAL = 30

	REQUEST_TIMEOUT = 300

	MAX_TXN_NUMBER_ONE_TIME = 128
)

type Registry interface {
	Err() <-chan error
	Ready() <-chan int
	PutNoOverride(ctx context.Context, opts ...PluginOpOption) (bool, error)
	Do(ctx context.Context, opts ...PluginOpOption) (*PluginResponse, error)
	Txn(ctx context.Context, ops []PluginOp) (*PluginResponse, error)
	TxnWithCmp(ctx context.Context, success []PluginOp, cmp []CompareOp, fail []PluginOp) (*PluginResponse, error)
	LeaseGrant(ctx context.Context, TTL int64) (leaseID int64, err error)
	LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error)
	LeaseRevoke(ctx context.Context, leaseID int64) error
	// this function block util:
	// 1. connection error
	// 2. call send function failed
	// 3. response.Err()
	// 4. time out to watch, but return nil
	Watch(ctx context.Context, opts ...PluginOpOption) error
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
	Action        ActionType    `json:"action"`
	Key           []byte        `json:"key,omitempty"`
	EndKey        []byte        `json:"endKey,omitempty"`
	Value         []byte        `json:"value,omitempty"`
	Prefix        bool          `json:"prefix,omitempty"`
	PrevKV        bool          `json:"prevKV,omitempty"`
	Lease         int64         `json:"leaseId,omitempty"`
	KeyOnly       bool          `json:"keyOnly,omitempty"`
	CountOnly     bool          `json:"countOnly,omitempty"`
	SortOrder     SortOrder     `json:"sort,omitempty"`
	Revision      int64         `json:"rev,omitempty"`
	IgnoreLease   bool          `json:"ignoreLease,omitempty"`
	Mode          CacheMode     `json:"mode,omitempty"`
	WatchCallback WatchCallback `json:"watchCallback,omitempty"`
}

func (op *PluginOp) String() string {
	b, _ := json.Marshal(op)
	return util.BytesToStringWithNoCopy(b)
}

type PluginOpOption func(*PluginOp)
type WatchCallback func(message string, evt *PluginResponse) error

var GET PluginOpOption = func(op *PluginOp) { op.Action = Get }
var PUT PluginOpOption = func(op *PluginOp) { op.Action = Put }
var DEL PluginOpOption = func(op *PluginOp) { op.Action = Delete }

func WithKey(key []byte) PluginOpOption      { return func(op *PluginOp) { op.Key = key } }
func WithEndKey(key []byte) PluginOpOption   { return func(op *PluginOp) { op.EndKey = key } }
func WithValue(value []byte) PluginOpOption  { return func(op *PluginOp) { op.Value = value } }
func WithPrefix() PluginOpOption             { return func(op *PluginOp) { op.Prefix = true } }
func WithPrevKv() PluginOpOption             { return func(op *PluginOp) { op.PrevKV = true } }
func WithLease(leaseID int64) PluginOpOption { return func(op *PluginOp) { op.Lease = leaseID } }
func WithKeyOnly() PluginOpOption            { return func(op *PluginOp) { op.KeyOnly = true } }
func WithCountOnly() PluginOpOption          { return func(op *PluginOp) { op.CountOnly = true } }
func WithNoneOrder() PluginOpOption          { return func(op *PluginOp) { op.SortOrder = SORT_NONE } }
func WithAscendOrder() PluginOpOption        { return func(op *PluginOp) { op.SortOrder = SORT_ASCEND } }
func WithDescendOrder() PluginOpOption       { return func(op *PluginOp) { op.SortOrder = SORT_DESCEND } }
func WithRev(revision int64) PluginOpOption  { return func(op *PluginOp) { op.Revision = revision } }
func WithIgnoreLease() PluginOpOption        { return func(op *PluginOp) { op.IgnoreLease = true } }
func WithCacheOnly() PluginOpOption          { return func(op *PluginOp) { op.Mode = MODE_CACHE } }
func WithNoCache() PluginOpOption            { return func(op *PluginOp) { op.Mode = MODE_NO_CACHE } }
func WithWatchCallback(f WatchCallback) PluginOpOption {
	return func(op *PluginOp) { op.WatchCallback = f }
}
func WithStrKey(key string) PluginOpOption     { return WithKey(util.StringToBytesWithNoCopy(key)) }
func WithStrEndKey(key string) PluginOpOption  { return WithEndKey(util.StringToBytesWithNoCopy(key)) }
func WithStrValue(value string) PluginOpOption { return WithValue(util.StringToBytesWithNoCopy(value)) }

func WatchPrefixOpOptions(key string) []PluginOpOption {
	return []PluginOpOption{GET, WithStrKey(key), WithPrefix(), WithPrevKv()}
}

func OpGet(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Get
	return
}
func OpPut(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Put
	return
}
func OpDel(opts ...PluginOpOption) (op PluginOp) {
	op = OptionsToOp(opts...)
	op.Action = Delete
	return
}
func OptionsToOp(opts ...PluginOpOption) (op PluginOp) {
	for _, opt := range opts {
		opt(&op)
	}
	return
}

type PluginResponse struct {
	Action    ActionType
	Kvs       []*mvccpb.KeyValue
	Count     int64
	Revision  int64
	Succeeded bool
}

func (pr *PluginResponse) String() string {
	return fmt.Sprintf("{action: %s, count: %d/%d, rev: %d, succeed: %v}",
		pr.Action, len(pr.Kvs), pr.Count, pr.Revision, pr.Succeeded)
}

type CompareOp struct {
	Key    []byte
	Type   CompareType
	Result CompareResult
	Value  interface{}
}

func CmpVer(key []byte) CompareOp          { return CompareOp{Key: key, Type: CMP_VERSION} }
func CmpCreateRev(key []byte) CompareOp    { return CompareOp{Key: key, Type: CMP_CREATE} }
func CmpModRev(key []byte) CompareOp       { return CompareOp{Key: key, Type: CMP_MOD} }
func CmpVal(key []byte) CompareOp          { return CompareOp{Key: key, Type: CMP_VALUE} }
func CmpStrVer(key string) CompareOp       { return CmpVer(util.StringToBytesWithNoCopy(key)) }
func CmpStrCreateRev(key string) CompareOp { return CmpCreateRev(util.StringToBytesWithNoCopy(key)) }
func CmpStrModRev(key string) CompareOp    { return CmpModRev(util.StringToBytesWithNoCopy(key)) }
func CmpStrVal(key string) CompareOp       { return CmpVal(util.StringToBytesWithNoCopy(key)) }
func OpCmp(cmp CompareOp, result CompareResult, v interface{}) CompareOp {
	cmp.Result = result
	cmp.Value = v
	return cmp
}

func init() {
	RegistryPlugins = make(map[string]func(cfg *Config) Registry)
}

var noClientPluginErr = fmt.Errorf("register center client plugin does not exist")

type ErrorRegisterCenterClient struct {
	ready chan int
}

func (ec *ErrorRegisterCenterClient) safeClose(chan int) {
	defer util.RecoverAndReport()
	close(ec.ready)
}
func (ec *ErrorRegisterCenterClient) Err() (err <-chan error) {
	return
}
func (ec *ErrorRegisterCenterClient) Ready() <-chan int {
	ec.safeClose(ec.ready)
	return ec.ready
}
func (ec *ErrorRegisterCenterClient) PutNoOverride(ctx context.Context, opts ...PluginOpOption) (bool, error) {
	return false, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) Do(ctx context.Context, opts ...PluginOpOption) (*PluginResponse, error) {
	return nil, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) Txn(ctx context.Context, ops []PluginOp) (*PluginResponse, error) {
	return nil, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) TxnWithCmp(ctx context.Context, success []PluginOp, cmp []CompareOp, fail []PluginOp) (*PluginResponse, error) {
	return nil, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) LeaseGrant(ctx context.Context, TTL int64) (leaseID int64, err error) {
	return 0, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) LeaseRenew(ctx context.Context, leaseID int64) (TTL int64, err error) {
	return 0, noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) LeaseRevoke(ctx context.Context, leaseID int64) error {
	return noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) Watch(ctx context.Context, opts ...PluginOpOption) error {
	return noClientPluginErr
}
func (ec *ErrorRegisterCenterClient) Close() {
	ec.safeClose(ec.ready)
}
func (ec *ErrorRegisterCenterClient) CompactCluster(ctx context.Context) {}
func (ec *ErrorRegisterCenterClient) Compact(ctx context.Context, revision int64) error {
	return noClientPluginErr
}

func RegisterCenterClientPlugin() func(cfg *Config) Registry {
	registryFunc, ok := RegistryPlugins[beego.AppConfig.String("registry_plugin")]
	if !ok {
		return func(*Config) Registry {
			return &ErrorRegisterCenterClient{
				ready: make(chan int),
			}
		}
	}
	return registryFunc
}

func RegisterCenterClient() (Registry, error) {
	registryFunc := RegisterCenterClientPlugin()
	autoSyncInterval, _ := beego.AppConfig.Int64("auto_sync_interval")
	if autoSyncInterval <= 0 {
		autoSyncInterval = REFRESH_MANAGER_CLUSTER_INTERVAL
	}
	instance := registryFunc(&Config{
		ClusterAddresses: beego.AppConfig.String("manager_cluster"),
		AutoSyncInterval: autoSyncInterval,
	})
	select {
	case err := <-instance.Err():
		return nil, err
	case <-instance.Ready():
	}
	return instance, nil
}

func GetRegisterCenter() Registry {
	if registryInstance == nil {
		singletonLock.Lock()
		for i := 0; registryInstance == nil; i++ {
			inst, err := RegisterCenterClient()
			if err != nil {
				util.Logger().Errorf(err, "get register center client failed")
			}
			registryInstance = inst

			if registryInstance != nil {
				singletonLock.Unlock()
				return registryInstance
			}

			if i >= len(wait_delay) {
				i = len(wait_delay) - 1
			}
			t := time.Duration(wait_delay[i]) * time.Second
			util.Logger().Errorf(nil, "initialize service center failed, retry after %s", t)
			<-time.After(t)
		}
		singletonLock.Unlock()
	}
	return registryInstance
}

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, REQUEST_TIMEOUT*time.Second)
}

func BatchCommit(ctx context.Context, opts []PluginOp) error {
	lenOpts := len(opts)
	tmpLen := lenOpts
	tmpOpts := []PluginOp{}
	var err error
	for i := 0; tmpLen > 0; i++ {
		tmpLen = lenOpts - (i+1)*MAX_TXN_NUMBER_ONE_TIME
		if tmpLen > 0 {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : (i+1)*MAX_TXN_NUMBER_ONE_TIME]
		} else {
			tmpOpts = opts[i*MAX_TXN_NUMBER_ONE_TIME : lenOpts]
		}
		_, err = GetRegisterCenter().Txn(ctx, tmpOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
