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
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

var defaultRegistryConfig Config

func init() {
	defaultRegistryConfig.ClusterAddresses = beego.AppConfig.DefaultString("manager_cluster", "sc-0=http://127.0.0.1:2380")
}

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

func (ct CompareType) String() string {
	switch ct {
	case CMP_VERSION:
		return "CMP_VERSION"
	case CMP_CREATE:
		return "CMP_CREATE"
	case CMP_MOD:
		return "CMP_MOD"
	case CMP_VALUE:
		return "CMP_VALUE"
	default:
		return "CMP_TYPE" + strconv.Itoa(int(ct))
	}
}

type CompareResult int

func (cr CompareResult) String() string {
	switch cr {
	case CMP_EQUAL:
		return "CMP_EQUAL"
	case CMP_GREATER:
		return "CMP_GREATER"
	case CMP_LESS:
		return "CMP_LESS"
	case CMP_NOT_EQUAL:
		return "CMP_NOT_EQUAL"
	default:
		return "CMP_RESULT" + strconv.Itoa(int(cr))
	}
}

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
	REQUEST_TIMEOUT = 300

	DEFAULT_PAGE_COUNT = 4096 // grpc does not allow to transport a large body more then 4MB in a request.
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
}

type Config struct {
	EmbedMode        string
	ClusterAddresses string
}

type PluginOp struct {
	Action        ActionType
	Key           []byte
	EndKey        []byte
	Value         []byte
	Prefix        bool
	PrevKV        bool
	Lease         int64
	KeyOnly       bool
	CountOnly     bool
	SortOrder     SortOrder
	Revision      int64
	IgnoreLease   bool
	Mode          CacheMode
	WatchCallback WatchCallback
	Offset        int64
	Limit         int64
}

func (op PluginOp) String() string {
	return fmt.Sprintf(
		"{mode: %s, action: %s, key: %s, end: %s, val: %d, prefix: %t, prev: %t, lease: %d, keyOnly: %t, countOnly: %t, sort: %s, rev: %d, ignoreLease: %t, offset: %d, limit: %d}",
		op.Mode, op.Action, op.Key, op.EndKey, len(util.BytesToStringWithNoCopy(op.Value)),
		op.Prefix, op.PrevKV, op.Lease, op.KeyOnly, op.CountOnly,
		op.SortOrder, op.Revision, op.IgnoreLease, op.Offset, op.Limit,
	)
}

type Operation func(...PluginOpOption) (op PluginOp)

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
func WithOffset(i int64) PluginOpOption        { return func(op *PluginOp) { op.Offset = i } }
func WithLimit(i int64) PluginOpOption         { return func(op *PluginOp) { op.Limit = i } }
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
	if op.Limit == 0 {
		op.Offset = -1
		op.Limit = DEFAULT_PAGE_COUNT
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

func (op CompareOp) String() string {
	return fmt.Sprintf(
		"{key: %s, type: %s, result: %s, val: %s}",
		op.Key, op.Type, op.Result, op.Value,
	)
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

func WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, REQUEST_TIMEOUT*time.Second)
}

func RegistryConfig() *Config {
	return &defaultRegistryConfig
}
