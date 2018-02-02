/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package etcd

import (
	"crypto/tls"
	"errors"
	"fmt"
	errorsEx "github.com/apache/incubator-servicecomb-service-center/pkg/errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	sctls "github.com/apache/incubator-servicecomb-service-center/server/tls"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"strings"
	"time"
)

const (
	CONNECT_MANAGER_SERVER_TIMEOUT = 10
)

var clientTLSConfig *tls.Config

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", NewRegistry})
}

type EtcdClient struct {
	Client *clientv3.Client
	err    chan error
	ready  chan int
}

func (s *EtcdClient) Err() <-chan error {
	return s.err
}

func (s *EtcdClient) Ready() <-chan int {
	return s.ready
}

func (s *EtcdClient) Close() {
	if s.Client != nil {
		s.Client.Close()
	}
	util.Logger().Debugf("etcd client stopped.")
}

func (c *EtcdClient) Compact(ctx context.Context, reserve int64) error {
	eps := c.Client.Endpoints()
	curRev := c.getLeaderCurrentRevision(ctx)

	revToCompact := max(0, curRev-reserve)
	if revToCompact <= 0 {
		util.Logger().Infof("revision is %d, <=%d, no nead to compact %s", curRev, reserve, eps)
		return nil
	}

	t := time.Now()
	_, err := c.Client.Compact(ctx, revToCompact, clientv3.WithCompactPhysical())
	if err != nil {
		util.Logger().Errorf(err, "Compact %s failed, revision is %d(current: %d, reserve %d)",
			eps, revToCompact, curRev, reserve)
		return err
	}
	util.LogInfoOrWarnf(t, "Compacted %s, revision is %d(current: %d, reserve %d)", eps, revToCompact, curRev, reserve)

	// TODO can not defrag! because backend will always be unavailable when space in used is too large.
	/*for _, ep := range eps {
		t = time.Now()
		_, err := c.Client.Defragment(ctx, ep)
		if err != nil {
			util.Logger().Errorf(err, "Defrag %s failed", ep)
			continue
		}
		util.LogInfoOrWarnf(t, "Defraged %s", ep)
	}*/

	return nil
}

func (c *EtcdClient) getLeaderCurrentRevision(ctx context.Context) int64 {
	eps := c.Client.Endpoints()
	curRev := int64(0)
	for _, ep := range eps {
		resp, err := c.Client.Status(ctx, ep)
		if err != nil {
			util.Logger().Error(fmt.Sprintf("Compact error ,can not get status from %s", ep), err)
			continue
		}
		curRev = resp.Header.Revision
		if resp.Leader == resp.Header.MemberId {
			util.Logger().Infof("Get leader endpoint: %s, revision is %d", ep, curRev)
			break
		}
	}
	return curRev
}

func max(n1, n2 int64) int64 {
	if n1 > n2 {
		return n1
	}
	return n2
}

func (s *EtcdClient) toGetRequest(op registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.Prefix {
		opts = append(opts, clientv3.WithPrefix())
	} else if len(op.EndKey) > 0 {
		opts = append(opts, clientv3.WithRange(util.BytesToStringWithNoCopy(op.EndKey)))
	}
	if op.PrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	if op.KeyOnly {
		opts = append(opts, clientv3.WithKeysOnly())
	}
	if op.CountOnly {
		opts = append(opts, clientv3.WithCountOnly())
	}
	if op.Revision > 0 {
		opts = append(opts, clientv3.WithRev(op.Revision))
	}
	switch op.SortOrder {
	case registry.SORT_ASCEND:
		opts = append(opts, clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	case registry.SORT_DESCEND:
		opts = append(opts, clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	}
	return opts
}

func (s *EtcdClient) toPutRequest(op registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.PrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	if op.Lease > 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(op.Lease)))
	}
	if op.IgnoreLease {
		// TODO WithIgnoreLease support
	}
	return opts
}

func (s *EtcdClient) toDeleteRequest(op registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.Prefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	if op.PrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	return opts
}

func (c *EtcdClient) toTxnRequest(opts []registry.PluginOp) []clientv3.Op {
	etcdOps := []clientv3.Op{}
	for _, op := range opts {
		switch op.Action {
		case registry.Get:
			etcdOps = append(etcdOps, clientv3.OpGet(util.BytesToStringWithNoCopy(op.Key), c.toGetRequest(op)...))
		case registry.Put:
			var value string = ""
			if len(op.Value) > 0 {
				value = util.BytesToStringWithNoCopy(op.Value)
			}
			etcdOps = append(etcdOps, clientv3.OpPut(util.BytesToStringWithNoCopy(op.Key), value, c.toPutRequest(op)...))
		case registry.Delete:
			etcdOps = append(etcdOps, clientv3.OpDelete(util.BytesToStringWithNoCopy(op.Key), c.toDeleteRequest(op)...))
		}
	}
	return etcdOps
}

func (c *EtcdClient) toCompares(cmps []registry.CompareOp) []clientv3.Cmp {
	etcdCmps := []clientv3.Cmp{}
	for _, cmp := range cmps {
		var cmpType clientv3.Cmp
		var cmpResult string
		key := util.BytesToStringWithNoCopy(cmp.Key)
		switch cmp.Type {
		case registry.CMP_VERSION:
			cmpType = clientv3.Version(key)
		case registry.CMP_CREATE:
			cmpType = clientv3.CreateRevision(key)
		case registry.CMP_MOD:
			cmpType = clientv3.ModRevision(key)
		case registry.CMP_VALUE:
			cmpType = clientv3.Value(key)
		}
		switch cmp.Result {
		case registry.CMP_EQUAL:
			cmpResult = "="
		case registry.CMP_GREATER:
			cmpResult = ">"
		case registry.CMP_LESS:
			cmpResult = "<"
		case registry.CMP_NOT_EQUAL:
			cmpResult = "!="
		}
		etcdCmps = append(etcdCmps, clientv3.Compare(cmpType, cmpResult, cmp.Value))
	}
	return etcdCmps
}

func (c *EtcdClient) PutNoOverride(ctx context.Context, opts ...registry.PluginOpOption) (bool, error) {
	op := registry.OpPut(opts...)
	resp, err := c.TxnWithCmp(ctx, []registry.PluginOp{op}, []registry.CompareOp{
		registry.OpCmp(registry.CmpCreateRev(op.Key), registry.CMP_EQUAL, 0),
	}, nil)
	if err != nil {
		util.Logger().Errorf(err, "PutNoOverride %s failed", op.Key)
		return false, err
	}
	util.Logger().Debugf("response %s %v %v", op.Key, resp.Succeeded, resp.Revision)
	return resp.Succeeded, nil
}

func (c *EtcdClient) paging(ctx context.Context, op registry.PluginOp) (*clientv3.GetResponse, error) {
	var etcdResp *clientv3.GetResponse
	key := util.BytesToStringWithNoCopy(op.Key)

	start := time.Now()
	tempOp := op
	tempOp.CountOnly = true
	coutResp, err := c.Client.Get(ctx, key, c.toGetRequest(tempOp)...)
	if err != nil {
		return nil, err
	}

	recordCount := coutResp.Count
	if op.Offset == -1 && recordCount < op.Limit {
		return nil, nil // no paging
	}

	tempOp.KeyOnly = false
	tempOp.CountOnly = false
	tempOp.Prefix = false
	tempOp.SortOrder = registry.SORT_ASCEND
	tempOp.EndKey = util.StringToBytesWithNoCopy(clientv3.GetPrefixRangeEnd(key))
	if len(op.EndKey) > 0 {
		tempOp.EndKey = op.EndKey
	}
	tempOp.Revision = coutResp.Header.Revision

	etcdResp = coutResp
	etcdResp.Kvs = make([]*mvccpb.KeyValue, 0, etcdResp.Count)

	pageCount := recordCount / op.Limit
	remainCount := recordCount % op.Limit
	if remainCount > 0 {
		pageCount++
	}

	baseOps := []clientv3.OpOption{}
	baseOps = append(baseOps, c.toGetRequest(tempOp)...)

	nextKey := key
	for i := int64(0); i < pageCount; i++ {
		limit := op.Limit
		if remainCount > 0 && i == pageCount-1 {
			limit = remainCount
		}
		ops := append(baseOps, clientv3.WithLimit(int64(limit)))
		recordResp, err := c.Client.Get(ctx, nextKey, ops...)
		if err != nil {
			return nil, err
		}
		l := int64(len(recordResp.Kvs))
		nextKey = clientv3.GetPrefixRangeEnd(util.BytesToStringWithNoCopy(recordResp.Kvs[l-1].Key))

		if op.Offset >= 0 {
			if op.Offset < i*op.Limit {
				continue
			} else if op.Offset >= (i+1)*op.Limit {
				break
			}
		}
		etcdResp.Kvs = append(etcdResp.Kvs, recordResp.Kvs...)
	}

	if op.Offset == -1 {
		util.LogInfoOrWarnf(start, "get too many KeyValues(%s) from etcdserver, now paging.(%d vs %d)",
			key, recordCount, op.Limit)
	}

	// too slow
	if op.SortOrder == registry.SORT_DESCEND {
		t := time.Now()
		for i, l := 0, len(etcdResp.Kvs); i < l; i++ {
			last := l - i - 1
			if last <= i {
				break
			}
			etcdResp.Kvs[i], etcdResp.Kvs[last] = etcdResp.Kvs[last], etcdResp.Kvs[i]
		}
		util.LogNilOrWarnf(t, "sorted descend %d KeyValues(%s)", recordCount, key)
	}
	return etcdResp, nil
}

func (c *EtcdClient) Do(ctx context.Context, opts ...registry.PluginOpOption) (*registry.PluginResponse, error) {
	var (
		err  error
		resp *registry.PluginResponse
	)

	start := time.Now()
	op := registry.OptionsToOp(opts...)

	span := TracingBegin(ctx, "etcd:do", op)
	defer TracingEnd(span, err)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()

	switch op.Action {
	case registry.Get:
		var etcdResp *clientv3.GetResponse
		key := util.BytesToStringWithNoCopy(op.Key)

		if (op.Prefix || len(op.EndKey) > 0) && !op.CountOnly {
			etcdResp, err = c.paging(ctx, op)
			if err != nil {
				break
			}
		}

		if etcdResp == nil {
			etcdResp, err = c.Client.Get(otCtx, key, c.toGetRequest(op)...)
			if err != nil {
				break
			}
		}

		resp = &registry.PluginResponse{
			Kvs:      etcdResp.Kvs,
			Count:    etcdResp.Count,
			Revision: etcdResp.Header.Revision,
		}
	case registry.Put:
		var value string = ""
		if len(op.Value) > 0 {
			value = util.BytesToStringWithNoCopy(op.Value)
		}
		var etcdResp *clientv3.PutResponse
		etcdResp, err = c.Client.Put(otCtx, util.BytesToStringWithNoCopy(op.Key), value, c.toPutRequest(op)...)
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Revision: etcdResp.Header.Revision,
		}
	case registry.Delete:
		var etcdResp *clientv3.DeleteResponse
		etcdResp, err = c.Client.Delete(otCtx, util.BytesToStringWithNoCopy(op.Key), c.toDeleteRequest(op)...)
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Revision: etcdResp.Header.Revision,
		}
	}

	if err != nil {
		return nil, err
	}

	resp.Succeeded = true

	util.LogNilOrWarnf(start, "registry client do %s", op)
	return resp, nil
}

func (c *EtcdClient) Txn(ctx context.Context, opts []registry.PluginOp) (*registry.PluginResponse, error) {
	resp, err := c.TxnWithCmp(ctx, opts, nil, nil)
	if err != nil {
		return nil, err
	}
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Revision,
	}, nil
}

func (c *EtcdClient) TxnWithCmp(ctx context.Context, success []registry.PluginOp, cmps []registry.CompareOp, fail []registry.PluginOp) (*registry.PluginResponse, error) {
	var err error

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()

	start := time.Now()
	etcdCmps := c.toCompares(cmps)
	etcdSuccessOps := c.toTxnRequest(success)
	etcdFailOps := c.toTxnRequest(fail)

	span := TracingBegin(ctx, "etcd:txn", success[0])
	defer TracingEnd(span, err)

	kvc := clientv3.NewKV(c.Client)
	txn := kvc.Txn(otCtx)
	if len(etcdCmps) > 0 {
		txn.If(etcdCmps...)
	}
	txn.Then(etcdSuccessOps...)
	if len(etcdFailOps) > 0 {
		txn.Else(etcdFailOps...)
	}
	resp, err := txn.Commit()
	if err != nil {
		return nil, err
	}
	util.LogNilOrWarnf(start, "registry client txn {if: %s, then: %d, else: %d}", cmps, len(success), len(fail))
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Header.Revision,
	}, nil
}

func (c *EtcdClient) LeaseGrant(ctx context.Context, TTL int64) (int64, error) {
	var err error
	span := TracingBegin(ctx, "etcd:grant",
		registry.PluginOp{Action: registry.Put, Key: util.StringToBytesWithNoCopy(fmt.Sprint(TTL))})
	defer TracingEnd(span, err)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	start := time.Now()
	etcdResp, err := c.Client.Grant(otCtx, TTL)
	if err != nil {
		return 0, err
	}
	util.LogNilOrWarnf(start, "registry client grant lease %ds", TTL)
	return int64(etcdResp.ID), nil
}

func (c *EtcdClient) LeaseRenew(ctx context.Context, leaseID int64) (int64, error) {
	var err error
	span := TracingBegin(ctx, "etcd:keepalive",
		registry.PluginOp{Action: registry.Put, Key: util.StringToBytesWithNoCopy(fmt.Sprint(leaseID))})
	defer TracingEnd(span, err)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	start := time.Now()
	etcdResp, err := c.Client.KeepAliveOnce(otCtx, clientv3.LeaseID(leaseID))
	if err != nil {
		if err.Error() == grpc.ErrorDesc(rpctypes.ErrGRPCLeaseNotFound) {
			return 0, err
		}
		return 0, errorsEx.RaiseError(err)
	}
	util.LogNilOrWarnf(start, "registry client renew lease %d", leaseID)
	return etcdResp.TTL, nil
}

func (c *EtcdClient) LeaseRevoke(ctx context.Context, leaseID int64) error {
	var err error
	span := TracingBegin(ctx, "etcd:revoke",
		registry.PluginOp{Action: registry.Delete, Key: util.StringToBytesWithNoCopy(fmt.Sprint(leaseID))})
	defer TracingEnd(span, err)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	start := time.Now()
	_, err = c.Client.Revoke(otCtx, clientv3.LeaseID(leaseID))
	if err != nil {
		if err.Error() == grpc.ErrorDesc(rpctypes.ErrGRPCLeaseNotFound) {
			return err
		}
		return errorsEx.RaiseError(err)
	}
	util.LogNilOrWarnf(start, "registry client revoke lease %d", leaseID)
	return nil
}

func (c *EtcdClient) Watch(ctx context.Context, opts ...registry.PluginOpOption) (err error) {
	op := registry.OpGet(opts...)

	n := len(op.Key)
	if n > 0 {
		// 必须创建新的client连接
		/*client, err := newClient(c.Client.Endpoints())
		  if err != nil {
		          util.Logger().Error("get manager client failed", err)
		          return err
		  }
		  defer client.Close()*/
		// 现在跟ETCD仅使用一个连接，共用client即可
		client := clientv3.NewWatcher(c.Client)
		defer client.Close()

		key := util.BytesToStringWithNoCopy(op.Key)
		if op.Prefix && key[len(key)-1] != '/' {
			key += "/"
		}

		// 不能设置超时context，内部判断了连接超时和watch超时
		ws := client.Watch(context.Background(), key, c.toGetRequest(op)...)

		var ok bool
		var resp clientv3.WatchResponse
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok = <-ws:
				if !ok {
					err = errors.New("channel is closed")
					return
				}
				// cause a rpc ResourceExhausted error if watch response body larger then 4MB
				if err = resp.Err(); err != nil {
					return
				}

				err = dispatch(resp.Events, op.WatchCallback)
				if err != nil {
					return
				}
			}
		}
	}
	return fmt.Errorf("no key has been watched")
}

func dispatch(evts []*clientv3.Event, cb registry.WatchCallback) error {
	l := len(evts)
	kvs := make([]*mvccpb.KeyValue, l)
	sIdx, eIdx, prevAction := 0, 0, mvccpb.PUT
	pResp := &registry.PluginResponse{Action: registry.Put, Succeeded: true}

	for _, evt := range evts {
		if prevAction != evt.Type {
			prevAction = evt.Type

			if eIdx > 0 {
				err := setResponseAndCallback(pResp, kvs[sIdx:eIdx], cb)
				if err != nil {
					return err
				}
				sIdx = eIdx
			}
		}

		if pResp.Revision < evt.Kv.ModRevision {
			pResp.Revision = evt.Kv.ModRevision
		}
		pResp.Action = setKvsAndConvertAction(kvs, eIdx, evt)

		eIdx++
	}

	if eIdx > 0 {
		return setResponseAndCallback(pResp, kvs[sIdx:eIdx], cb)
	}
	return nil
}

func setKvsAndConvertAction(kvs []*mvccpb.KeyValue, pIdx int, evt *clientv3.Event) registry.ActionType {
	switch evt.Type {
	case mvccpb.DELETE:
		kv := evt.PrevKv
		if kv == nil {
			kv = evt.Kv
		}
		kvs[pIdx] = kv
		return registry.Delete
	default:
		kvs[pIdx] = evt.Kv
		return registry.Put
	}
}

func setResponseAndCallback(pResp *registry.PluginResponse, kvs []*mvccpb.KeyValue, cb registry.WatchCallback) error {
	pResp.Count = int64(len(kvs))
	pResp.Kvs = kvs
	return cb("key information changed", pResp)
}

func NewRegistry() mgr.PluginInstance {
	util.Logger().Warnf(nil, "starting service center in proxy mode")

	inst := &EtcdClient{
		err:   make(chan error, 1),
		ready: make(chan int),
	}
	addrs := strings.Split(registry.RegistryConfig().ClusterAddresses, ",")

	if core.ServerInfo.Config.SslEnabled && strings.Index(registry.RegistryConfig().ClusterAddresses, "https://") >= 0 {
		var err error
		// go client tls限制，提供身份证书、不认证服务端、不校验CN
		clientTLSConfig, err = sctls.GetClientTLSConfig()
		if err != nil {
			util.Logger().Error("get etcd client tls config failed", err)
			inst.err <- err
			return inst
		}
	}

	endpoints := []string{}
	for _, addr := range addrs {
		if strings.Index(addr, "://") > 0 {
			// 如果配置格式为"sr-0=http(s)://IP:Port"，则需要分离IP:Port部分
			endpoints = append(endpoints, addr[strings.Index(addr, "://")+3:])
		} else {
			endpoints = append(endpoints, addr)
		}

	}

	inv, err := time.ParseDuration(core.ServerInfo.Config.AutoSyncInterval)
	if err != nil {
		util.Logger().Errorf(err, "invalid auto sync interval '%s'.", core.ServerInfo.Config.AutoSyncInterval)
	}
	client, err := newClient(endpoints, inv)
	if err != nil {
		util.Logger().Errorf(err, "get etcd client %v failed.", endpoints)
		inst.err <- err
		return inst
	}

	util.Logger().Warnf(nil, "get etcd client %v completed, auto sync endpoints interval is %s.",
		endpoints, core.ServerInfo.Config.AutoSyncInterval)
	inst.Client = client
	close(inst.ready)
	return inst
}

func newClient(endpoints []string, autoSyncInterval time.Duration) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      CONNECT_MANAGER_SERVER_TIMEOUT * time.Second,
		TLS:              clientTLSConfig, // 暂时与API Server共用一套证书
		AutoSyncInterval: autoSyncInterval,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
