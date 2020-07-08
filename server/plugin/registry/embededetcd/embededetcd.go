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

package embededetcd

import (
	"context"
	"errors"
	"fmt"
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/coreos/etcd/compactor"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/lease"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"net/url"
	"strings"
	"time"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{PName: mgr.REGISTRY, Name: "embeded_etcd", New: getEmbedInstance})
}

type EtcdEmbed struct {
	Embed     *embed.Etcd
	err       chan error
	ready     chan struct{}
	goroutine *gopool.Pool
}

func (s *EtcdEmbed) Err() <-chan error {
	return s.err
}

func (s *EtcdEmbed) Ready() <-chan struct{} {
	return s.ready
}

func (s *EtcdEmbed) Close() {
	if s.Embed != nil {
		s.Embed.Close()
	}
	s.goroutine.Close(true)
	log.Debugf("embedded etcd client stopped")
}

func (s *EtcdEmbed) getPrefixEndKey(prefix []byte) []byte {
	l := len(prefix)
	endBytes := make([]byte, l+1)
	copy(endBytes, prefix)
	if endBytes[l-1] == 0xff {
		endBytes[l] = 1
		return endBytes
	}
	endBytes[l-1]++
	return endBytes[:l]
}

func (s *EtcdEmbed) toGetRequest(op registry.PluginOp) *etcdserverpb.RangeRequest {
	endBytes := op.EndKey
	if op.Prefix {
		endBytes = s.getPrefixEndKey(op.Key)
	}
	order := etcdserverpb.RangeRequest_NONE
	switch op.SortOrder {
	case registry.SortAscend:
		order = etcdserverpb.RangeRequest_ASCEND
	case registry.SortDescend:
		order = etcdserverpb.RangeRequest_DESCEND
	}
	return &etcdserverpb.RangeRequest{
		Key:        op.Key,
		RangeEnd:   endBytes,
		KeysOnly:   op.KeyOnly,
		CountOnly:  op.CountOnly,
		SortOrder:  order,
		SortTarget: etcdserverpb.RangeRequest_KEY,
		Revision:   op.Revision,
	}
}

func (s *EtcdEmbed) toPutRequest(op registry.PluginOp) *etcdserverpb.PutRequest {
	var valueBytes []byte
	if len(op.Value) > 0 {
		valueBytes = op.Value
	}
	return &etcdserverpb.PutRequest{
		Key:         op.Key,
		Value:       valueBytes,
		PrevKv:      op.PrevKV,
		Lease:       op.Lease,
		IgnoreLease: op.IgnoreLease,
	}
}

func (s *EtcdEmbed) toDeleteRequest(op registry.PluginOp) *etcdserverpb.DeleteRangeRequest {
	endBytes := op.EndKey
	if op.Prefix {
		endBytes = s.getPrefixEndKey(op.Key)
	}
	return &etcdserverpb.DeleteRangeRequest{
		Key:      op.Key,
		RangeEnd: endBytes,
		PrevKv:   op.PrevKV,
	}
}

func (s *EtcdEmbed) toTxnRequest(opts []registry.PluginOp) []*etcdserverpb.RequestOp {
	etcdOps := []*etcdserverpb.RequestOp{}
	for _, op := range opts {
		switch op.Action {
		case registry.Get:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestRange{
					RequestRange: s.toGetRequest(op),
				},
			})
		case registry.Put:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestPut{
					RequestPut: s.toPutRequest(op),
				},
			})
		case registry.Delete:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestDeleteRange{
					RequestDeleteRange: s.toDeleteRequest(op),
				},
			})
		}
	}
	return etcdOps
}

func (s *EtcdEmbed) toCompares(cmps []registry.CompareOp) []*etcdserverpb.Compare {
	etcdCmps := []*etcdserverpb.Compare{}
	for _, cmp := range cmps {
		compare := &etcdserverpb.Compare{
			Key: cmp.Key,
		}
		switch cmp.Type {
		case registry.CmpVersion:
			var version int64
			if cmp.Value != nil {
				if v, ok := cmp.Value.(int64); ok {
					version = v
				}
			}
			compare.Target = etcdserverpb.Compare_VERSION
			compare.TargetUnion = &etcdserverpb.Compare_Version{
				Version: version,
			}
		case registry.CmpCreate:
			var revision int64
			if cmp.Value != nil {
				if v, ok := cmp.Value.(int64); ok {
					revision = v
				}
			}
			compare.Target = etcdserverpb.Compare_CREATE
			compare.TargetUnion = &etcdserverpb.Compare_CreateRevision{
				CreateRevision: revision,
			}
		case registry.CmpMod:
			var revision int64
			if cmp.Value != nil {
				if v, ok := cmp.Value.(int64); ok {
					revision = v
				}
			}
			compare.Target = etcdserverpb.Compare_MOD
			compare.TargetUnion = &etcdserverpb.Compare_ModRevision{
				ModRevision: revision,
			}
		case registry.CmpValue:
			var value []byte
			if cmp.Value != nil {
				if v, ok := cmp.Value.([]byte); ok {
					value = v
				}
			}
			compare.Target = etcdserverpb.Compare_VALUE
			compare.TargetUnion = &etcdserverpb.Compare_Value{
				Value: value,
			}
		}
		switch cmp.Result {
		case registry.CmpEqual:
			compare.Result = etcdserverpb.Compare_EQUAL
		case registry.CmpGreater:
			compare.Result = etcdserverpb.Compare_GREATER
		case registry.CmpLess:
			compare.Result = etcdserverpb.Compare_LESS
		case registry.CmpNotEqual:
			compare.Result = etcdserverpb.Compare_NOT_EQUAL
		}
		etcdCmps = append(etcdCmps, compare)
	}
	return etcdCmps
}

func (s *EtcdEmbed) Compact(ctx context.Context, reserve int64) error {
	curRev := s.getLeaderCurrentRevision(ctx)
	revToCompact := max(0, curRev-reserve)
	if revToCompact <= 0 {
		log.Infof("revision is %d, <=%d, no nead to compact", curRev, reserve)
		return nil
	}

	log.Infof("compacting... revision is %d(current: %d, reserve %d)", revToCompact, curRev, reserve)
	_, err := s.Embed.Server.Compact(ctx, &etcdserverpb.CompactionRequest{
		Revision: revToCompact,
		Physical: true,
	})
	if err != nil {
		log.Errorf(err, "compact locally failed, revision is %d(current: %d, reserve %d)",
			revToCompact, curRev, reserve)
		return err
	}
	log.Infof("compacted locally, revision is %d(current: %d, reserve %d)", revToCompact, curRev, reserve)

	// TODO defragment
	log.Infof("defraged locally")

	return nil
}

func (s *EtcdEmbed) getLeaderCurrentRevision(ctx context.Context) int64 {
	return s.Embed.Server.KV().Rev()
}

func (s *EtcdEmbed) PutNoOverride(ctx context.Context, opts ...registry.PluginOpOption) (bool, error) {
	op := registry.OpPut(opts...)
	resp, err := s.TxnWithCmp(ctx, []registry.PluginOp{op}, []registry.CompareOp{
		registry.OpCmp(registry.CmpCreateRev(op.Key), registry.CmpEqual, 0),
	}, nil)
	log.Debugf("response %s %v %v", op.Key, resp.Succeeded, resp.Revision)
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}

func (s *EtcdEmbed) Do(ctx context.Context, opts ...registry.PluginOpOption) (*registry.PluginResponse, error) {
	op := registry.OptionsToOp(opts...)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	var err error
	var resp *registry.PluginResponse
	switch op.Action {
	case registry.Get:
		var etcdResp *etcdserverpb.RangeResponse
		etcdResp, err = s.Embed.Server.Range(otCtx, s.toGetRequest(op))
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Kvs:      etcdResp.Kvs,
			Count:    etcdResp.Count,
			Revision: etcdResp.Header.Revision,
		}
	case registry.Put:
		var etcdResp *etcdserverpb.PutResponse
		etcdResp, err = s.Embed.Server.Put(otCtx, s.toPutRequest(op))
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Revision: etcdResp.Header.Revision,
		}
	case registry.Delete:
		var etcdResp *etcdserverpb.DeleteRangeResponse
		etcdResp, err = s.Embed.Server.DeleteRange(otCtx, s.toDeleteRequest(op))
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
	return resp, nil
}

// TODO EMBED支持KV().TxnBegin()->TxnID，可惜PROXY模式暂时不支持
func (s *EtcdEmbed) Txn(ctx context.Context, opts []registry.PluginOp) (*registry.PluginResponse, error) {
	resp, err := s.TxnWithCmp(ctx, opts, nil, nil)
	if err != nil {
		return nil, err
	}
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Revision,
	}, nil
}

func (s *EtcdEmbed) TxnWithCmp(ctx context.Context, success []registry.PluginOp, cmps []registry.CompareOp, fail []registry.PluginOp) (*registry.PluginResponse, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()

	etcdCmps := s.toCompares(cmps)
	etcdSuccessOps := s.toTxnRequest(success)
	etcdFailOps := s.toTxnRequest(fail)
	txnRequest := &etcdserverpb.TxnRequest{
		Success: etcdSuccessOps,
	}
	if len(etcdCmps) > 0 {
		txnRequest.Compare = etcdCmps
	}
	if len(etcdFailOps) > 0 {
		txnRequest.Failure = etcdFailOps
	}
	resp, err := s.Embed.Server.Txn(otCtx, txnRequest)
	if err != nil {
		if err.Error() == rpctypes.ErrKeyNotFound.Error() {
			// etcd return ErrKeyNotFound if key does not exist and
			// the PUT options contain WithIgnoreLease
			return &registry.PluginResponse{Succeeded: false}, nil
		}
		return nil, err
	}

	var rangeResponse etcdserverpb.RangeResponse
	for _, itf := range resp.Responses {
		if rr, ok := itf.Response.(*etcdserverpb.ResponseOp_ResponseRange); ok {
			// plz request the same type range kv in txn success/fail options
			rangeResponse.Kvs = append(rangeResponse.Kvs, rr.ResponseRange.Kvs...)
			rangeResponse.Count += rr.ResponseRange.Count
		}
	}

	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Header.Revision,
		Kvs:       rangeResponse.Kvs,
		Count:     rangeResponse.Count,
	}, nil
}

func (s *EtcdEmbed) LeaseGrant(ctx context.Context, TTL int64) (int64, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	etcdResp, err := s.Embed.Server.LeaseGrant(otCtx, &etcdserverpb.LeaseGrantRequest{
		TTL: TTL,
	})
	if err != nil {
		return 0, err
	}
	return etcdResp.ID, nil
}

func (s *EtcdEmbed) LeaseRenew(ctx context.Context, leaseID int64) (int64, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	ttl, err := s.Embed.Server.LeaseRenew(otCtx, lease.LeaseID(leaseID))
	if err != nil {
		if err.Error() == rpctypes.ErrLeaseNotFound.Error() {
			return 0, err
		}
		return 0, errorsEx.RaiseError(err)
	}
	return ttl, nil
}

func (s *EtcdEmbed) LeaseRevoke(ctx context.Context, leaseID int64) error {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	_, err := s.Embed.Server.LeaseRevoke(otCtx, &etcdserverpb.LeaseRevokeRequest{
		ID: leaseID,
	})
	if err != nil {
		if err.Error() == rpctypes.ErrLeaseNotFound.Error() {
			return err
		}
		return errorsEx.RaiseError(err)
	}
	return nil
}

func (s *EtcdEmbed) Watch(ctx context.Context, opts ...registry.PluginOpOption) (err error) {
	op := registry.OpGet(opts...)

	if len(op.Key) > 0 {
		watchable := s.Embed.Server.Watchable()
		ws := watchable.NewWatchStream()
		defer ws.Close()

		key := util.BytesToStringWithNoCopy(op.Key)
		var keyBytes []byte
		if op.Prefix {
			if key[len(key)-1] != '/' {
				key += "/"
			}
			keyBytes = s.getPrefixEndKey(util.StringToBytesWithNoCopy(key))
		}
		watchID := ws.Watch(op.Key, keyBytes, op.Revision)
		defer func() {
			if err := ws.Cancel(watchID); err != nil {
				log.Error("", err)
			}
		}()
		responses := ws.Chan()
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-responses:
				if !ok {
					err = errors.New("channel is closed")
					return
				}

				err = dispatch(resp.Events, op.WatchCallback)
				if err != nil {
					return
				}
			}
		}
	}
	err = fmt.Errorf("no key has been watched")
	return
}

func (s *EtcdEmbed) readyNotify() {
	timeout := registry.Configuration().DialTimeout
	select {
	case <-s.Embed.Server.ReadyNotify():
		close(s.ready)
		s.goroutine.Do(func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case err := <-s.Embed.Err():
				s.err <- err
			}
		})
	case <-time.After(timeout):
		err := fmt.Errorf("timed out(%s)", timeout)
		log.Errorf(err, "read notify failed")

		s.Embed.Server.Stop()

		s.err <- err
	}
}

func dispatch(evts []mvccpb.Event, cb registry.WatchCallback) error {
	l := len(evts)
	kvs := make([]*mvccpb.KeyValue, l)
	sIdx, eIdx, rev := 0, 0, int64(0)
	action, prevEvtType := registry.Put, mvccpb.PUT

	for _, evt := range evts {
		if prevEvtType != evt.Type {
			if eIdx > 0 {
				err := callback(action, rev, kvs[sIdx:eIdx], cb)
				if err != nil {
					return err
				}
				sIdx = eIdx
			}
			prevEvtType = evt.Type
		}

		if rev < evt.Kv.ModRevision {
			rev = evt.Kv.ModRevision
		}
		action = setKvsAndConvertAction(kvs, eIdx, evt)

		eIdx++
	}

	if eIdx > 0 {
		return callback(action, rev, kvs[sIdx:eIdx], cb)
	}
	return nil
}

func setKvsAndConvertAction(kvs []*mvccpb.KeyValue, pIdx int, evt mvccpb.Event) registry.ActionType {
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

func callback(action registry.ActionType, rev int64, kvs []*mvccpb.KeyValue, cb registry.WatchCallback) error {
	return cb("key information changed", &registry.PluginResponse{
		Action:    action,
		Kvs:       kvs,
		Count:     int64(len(kvs)),
		Revision:  rev,
		Succeeded: true,
	})
}

func getEmbedInstance() mgr.Instance {
	log.Warnf("enable embedded registry mode")

	hostName := "sc-0"
	if len(registry.Configuration().ClusterName) > 0 {
		hostName = registry.Configuration().ClusterName
	}
	mgrAddrs := "http://127.0.0.1:2380"
	if len(registry.Configuration().ManagerAddress) > 0 {
		mgrAddrs = registry.Configuration().ManagerAddress
	}

	inst := &EtcdEmbed{
		err:       make(chan error, 1),
		ready:     make(chan struct{}),
		goroutine: gopool.New(context.Background()),
	}

	if registry.Configuration().SslEnabled {
		log.Info("config no use for embedded etcd")
	}

	serverCfg := embed.NewConfig()
	serverCfg.EnableV2 = false
	serverCfg.EnablePprof = false
	serverCfg.QuotaBackendBytes = etcdserver.MaxQuotaBytes
	// TODO 不支持使用TLS通信
	// 存储目录，相对于工作目录
	serverCfg.Dir = "data"
	// 集群支持
	serverCfg.Name = hostName
	serverCfg.InitialCluster = registry.Configuration().ClusterAddresses
	// 1. 管理端口
	urls, err := parseURL(mgrAddrs)
	if err != nil {
		log.Error(`"manager_addr" field configure error`, err)
		inst.err <- err
		return inst
	}
	serverCfg.LPUrls = urls
	serverCfg.APUrls = urls
	// 2. 业务端口，关闭默认2379端口
	serverCfg.LCUrls = nil
	serverCfg.ACUrls = nil
	// 自动压缩历史, 1 hour
	serverCfg.AutoCompactionMode = compactor.ModePeriodic
	serverCfg.AutoCompactionRetention = "1h"

	log.Debugf("--initial-cluster %s --initial-advertise-peer-urls %s --listen-peer-urls %s",
		serverCfg.InitialCluster, mgrAddrs, mgrAddrs)

	etcd, err := embed.StartEtcd(serverCfg)
	if err != nil {
		log.Error("error to start etcd server", err)
		inst.err <- err
		return inst
	}
	inst.Embed = etcd

	inst.readyNotify()
	return inst
}

func parseURL(addrs string) ([]url.URL, error) {
	var urls []url.URL
	ips := strings.Split(addrs, ",")
	for _, ip := range ips {
		addr, err := url.Parse(ip)
		if err != nil {
			log.Error("Error to parse ip address string", err)
			return nil, err
		}
		urls = append(urls, *addr)
	}
	return urls, nil
}

func max(n1, n2 int64) int64 {
	if n1 > n2 {
		return n1
	}
	return n2
}
