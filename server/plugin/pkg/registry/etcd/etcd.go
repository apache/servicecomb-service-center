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
	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/alarm"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// here will new an etcd connection after about 30s(=5s * 3 + (backoff:8s))
	// when the connected etcd member was hung but tcp is still alive
	healthCheckTimeout    = 5 * time.Second
	healthCheckRetryTimes = 3
)

var firstEndpoint string

func init() {
	clientv3.SetLogger(&clientLogger{})
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", NewRegistry})
}

type EtcdClient struct {
	Client *clientv3.Client

	Endpoints        []string
	DialTimeout      time.Duration
	TLSConfig        *tls.Config
	AutoSyncInterval time.Duration

	err       chan error
	ready     chan struct{}
	goroutine *gopool.Pool
}

func (c *EtcdClient) Initialize() (err error) {
	c.err = make(chan error, 1)
	c.ready = make(chan struct{})
	c.goroutine = gopool.New(context.Background())

	if len(c.Endpoints) == 0 {
		// parse the endpoints from config
		c.parseEndpoints()
	}
	if c.TLSConfig == nil && registry.Configuration().SslEnabled {
		var err error
		// go client tls限制，提供身份证书、不认证服务端、不校验CN
		c.TLSConfig, err = mgr.Plugins().TLS().ClientConfig()
		if err != nil {
			log.Error("get etcd client tls config failed", err)
			return err
		}
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = registry.Configuration().DialTimeout
	}
	if c.AutoSyncInterval == 0 {
		c.AutoSyncInterval = registry.Configuration().AutoSyncInterval
	}

	c.Client, err = c.newClient()
	if err != nil {
		log.Errorf(err, "get etcd client %v failed.", c.Endpoints)
		return
	}

	c.HealthCheck()

	close(c.ready)

	log.Warnf("get etcd client %v completed, auto sync endpoints interval is %s.",
		c.Endpoints, c.AutoSyncInterval)
	return
}

func (c *EtcdClient) newClient() (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:        c.Endpoints,
		DialTimeout:      c.DialTimeout,
		TLS:              c.TLSConfig,
		AutoSyncInterval: 0,
	})
	defer func() {
		if err != nil {
			alarm.Raise(alarm.IdBackendConnectionRefuse, alarm.AdditionalContext("%v", err))

			if client != nil {
				client.Close()
			}
			return
		}
		alarm.Clear(alarm.IdBackendConnectionRefuse)
	}()

	if err != nil {
		return nil, err
	}

	ctx, _ := context.WithTimeout(client.Ctx(), healthCheckTimeout)
	resp, err := client.MemberList(ctx)
	if err != nil {
		return nil, err
	}

	ReportBackendInstance(len(resp.Members))

	if len(c.Endpoints) == 1 {
		// no need to check remote endpoints
		return client, nil
	}

epLoop:
	for _, ep := range c.Endpoints {
		var cluster []string
		for _, mem := range resp.Members {
			for _, curl := range mem.ClientURLs {
				u, err := url.Parse(curl)
				if err != nil {
					return nil, err
				}
				cluster = append(cluster, u.Host)
				if u.Host == ep {
					continue epLoop
				}
			}
		}
		// maybe endpoints = [domain A, domain B] or there are more than one cluster
		err = fmt.Errorf("the etcd cluster endpoint list%v does not contain %s", cluster, ep)
		return nil, err
	}

	return client, nil
}

func (c *EtcdClient) Err() <-chan error {
	return c.err
}

func (c *EtcdClient) Ready() <-chan struct{} {
	return c.ready
}

func (c *EtcdClient) Close() {
	c.goroutine.Close(true)

	if c.Client != nil {
		c.Client.Close()
	}
	log.Debugf("etcd client stopped")
}

func (c *EtcdClient) Compact(ctx context.Context, reserve int64) error {
	eps := c.Client.Endpoints()
	curRev := c.getLeaderCurrentRevision(ctx)

	revToCompact := max(0, curRev-reserve)
	if revToCompact <= 0 {
		log.Infof("revision is %d, <=%d, no nead to compact %s", curRev, reserve, eps)
		return nil
	}

	t := time.Now()
	_, err := c.Client.Compact(ctx, revToCompact, clientv3.WithCompactPhysical())
	if err != nil {
		log.Errorf(err, "compact %s failed, revision is %d(current: %d, reserve %d)",
			eps, revToCompact, curRev, reserve)
		return err
	}
	log.LogInfoOrWarnf(t, "compacted %s, revision is %d(current: %d, reserve %d)", eps, revToCompact, curRev, reserve)

	// TODO can not defrag! because backend will always be unavailable when space in used is too large.
	/*for _, ep := range eps {
		t = time.Now()
		_, err := c.Client.Defragment(ctx, ep)
		if err != nil {
			log.Errorf(err, "Defrag %s failed", ep)
			continue
		}
		log.LogInfoOrWarnf(t, "Defraged %s", ep)
	}*/

	return nil
}

func (c *EtcdClient) getLeaderCurrentRevision(ctx context.Context) int64 {
	eps := c.Client.Endpoints()
	curRev := int64(0)
	for _, ep := range eps {
		resp, err := c.Client.Status(ctx, ep)
		if err != nil {
			log.Error(fmt.Sprintf("compact error ,can not get status from %s", ep), err)
			continue
		}
		curRev = resp.Header.Revision
		if resp.Leader == resp.Header.MemberId {
			log.Infof("get leader endpoint: %s, revision is %d", ep, curRev)
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

func (c *EtcdClient) toGetRequest(op registry.PluginOp) []clientv3.OpOption {
	var opts []clientv3.OpOption
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

func (c *EtcdClient) toPutRequest(op registry.PluginOp) []clientv3.OpOption {
	var opts []clientv3.OpOption
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

func (c *EtcdClient) toDeleteRequest(op registry.PluginOp) []clientv3.OpOption {
	var opts []clientv3.OpOption
	if op.Prefix {
		opts = append(opts, clientv3.WithPrefix())
	} else if len(op.EndKey) > 0 {
		opts = append(opts, clientv3.WithRange(util.BytesToStringWithNoCopy(op.EndKey)))
	}
	if op.PrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	return opts
}

func (c *EtcdClient) toTxnRequest(opts []registry.PluginOp) []clientv3.Op {
	var etcdOps []clientv3.Op
	for _, op := range opts {
		switch op.Action {
		case registry.Get:
			etcdOps = append(etcdOps, clientv3.OpGet(util.BytesToStringWithNoCopy(op.Key), c.toGetRequest(op)...))
		case registry.Put:
			var value string
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
	var etcdCmps []clientv3.Cmp
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
		log.Errorf(err, "PutNoOverride %s failed", op.Key)
		return false, err
	}
	return resp.Succeeded, nil
}

func (c *EtcdClient) paging(ctx context.Context, op registry.PluginOp) (*clientv3.GetResponse, error) {
	var etcdResp *clientv3.GetResponse
	key := util.BytesToStringWithNoCopy(op.Key)

	start := time.Now()
	tempOp := op
	tempOp.CountOnly = true
	countResp, err := c.Client.Get(ctx, key, c.toGetRequest(tempOp)...)
	if err != nil {
		return nil, err
	}

	recordCount := countResp.Count
	if op.Offset == -1 && recordCount <= op.Limit {
		return nil, nil // no need to do paging
	}

	tempOp.CountOnly = false
	tempOp.Prefix = false
	tempOp.SortOrder = registry.SORT_ASCEND
	tempOp.EndKey = op.EndKey
	if len(op.EndKey) == 0 {
		tempOp.EndKey = util.StringToBytesWithNoCopy(clientv3.GetPrefixRangeEnd(key))
	}
	tempOp.Revision = countResp.Header.Revision

	etcdResp = countResp
	etcdResp.Kvs = make([]*mvccpb.KeyValue, 0, etcdResp.Count)

	pageCount := recordCount / op.Limit
	remainCount := recordCount % op.Limit
	if remainCount > 0 {
		pageCount++
	}
	minPage, maxPage := int64(0), pageCount
	if op.Offset >= 0 {
		count := op.Offset + 1
		maxPage = count / op.Limit
		if count%op.Limit > 0 {
			maxPage++
		}
		minPage = maxPage - 1
	}

	var baseOps []clientv3.OpOption
	baseOps = append(baseOps, c.toGetRequest(tempOp)...)

	nextKey := key
	for i := int64(0); i < pageCount; i++ {
		if i >= maxPage {
			break
		}

		limit, start := op.Limit, 0
		if remainCount > 0 && i == pageCount-1 {
			limit = remainCount
		}
		if i != 0 {
			limit += 1
			start = 1
		}
		ops := append(baseOps, clientv3.WithLimit(int64(limit)))
		recordResp, err := c.Client.Get(ctx, nextKey, ops...)
		if err != nil {
			return nil, err
		}

		l := int64(len(recordResp.Kvs))
		nextKey = util.BytesToStringWithNoCopy(recordResp.Kvs[l-1].Key)
		if i < minPage {
			// even through current page index less then the min page index,
			// but here must to get the nextKey and then continue
			continue
		}
		etcdResp.Kvs = append(etcdResp.Kvs, recordResp.Kvs[start:]...)
	}

	if op.Offset == -1 {
		log.LogInfoOrWarnf(start, "get too many KeyValues(%s) from etcd, now paging.(%d vs %d)",
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
		log.LogNilOrWarnf(t, "sorted descend %d KeyValues(%s)", recordCount, key)
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
		var value string
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

	log.LogNilOrWarnf(start, "registry client do %s", op)
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

	var traceOps []registry.PluginOp
	traceOps = append(traceOps, success...)
	traceOps = append(traceOps, fail...)
	if len(traceOps) == 0 {
		return nil, fmt.Errorf("requested success or fail PluginOp list")
	}

	span := TracingBegin(ctx, "etcd:txn", traceOps[0])
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
	log.LogNilOrWarnf(start, "registry client txn {if(%v): %s, then: %d, else: %d}, rev: %d",
		resp.Succeeded, cmps, len(success), len(fail), resp.Header.Revision)

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

func (c *EtcdClient) LeaseGrant(ctx context.Context, TTL int64) (int64, error) {
	var err error
	span := TracingBegin(ctx, "etcd:grant",
		registry.PluginOp{Action: registry.Put, Key: util.StringToBytesWithNoCopy(strconv.FormatInt(TTL, 10))})
	defer TracingEnd(span, err)

	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	start := time.Now()
	etcdResp, err := c.Client.Grant(otCtx, TTL)
	if err != nil {
		return 0, err
	}
	log.LogNilOrWarnf(start, "registry client grant lease %ds", TTL)
	return int64(etcdResp.ID), nil
}

func (c *EtcdClient) LeaseRenew(ctx context.Context, leaseID int64) (int64, error) {
	var err error
	span := TracingBegin(ctx, "etcd:keepalive",
		registry.PluginOp{Action: registry.Put, Key: util.StringToBytesWithNoCopy(strconv.FormatInt(leaseID, 10))})
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
	log.LogNilOrWarnf(start, "registry client renew lease %d", leaseID)
	return etcdResp.TTL, nil
}

func (c *EtcdClient) LeaseRevoke(ctx context.Context, leaseID int64) error {
	var err error
	span := TracingBegin(ctx, "etcd:revoke",
		registry.PluginOp{Action: registry.Delete, Key: util.StringToBytesWithNoCopy(strconv.FormatInt(leaseID, 10))})
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
	log.LogNilOrWarnf(start, "registry client revoke lease %d", leaseID)
	return nil
}

func (c *EtcdClient) Watch(ctx context.Context, opts ...registry.PluginOpOption) (err error) {
	op := registry.OpGet(opts...)

	n := len(op.Key)
	if n > 0 {
		client := clientv3.NewWatcher(c.Client)
		defer client.Close()

		key := util.BytesToStringWithNoCopy(op.Key)

		// 不能设置超时context，内部判断了连接超时和watch超时
		wCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// #9103: must be not a context.TODO/Background, because the WatchChan will not closed when finish.
		ws := client.Watch(wCtx, key, c.toGetRequest(op)...)

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

func (c *EtcdClient) HealthCheck() {
	if c.AutoSyncInterval >= time.Second {
		c.goroutine.Do(c.healthCheckLoop)
	}
}

func (c *EtcdClient) healthCheckLoop(pctx context.Context) {
	retries, start := healthCheckRetryTimes, time.Now()
hcLoop:
	for {
		select {
		case <-pctx.Done():
			return
		case <-time.After(c.AutoSyncInterval):
			var err error
			for i := 0; i < retries; i++ {
				ctx, _ := context.WithTimeout(c.Client.Ctx(), healthCheckTimeout)
				if err = c.SyncMembers(ctx); err != nil {
					d := util.GetBackoff().Delay(i)
					log.Errorf(err, "retry to sync members from etcd %s after %s", c.Endpoints, d)
					select {
					case <-pctx.Done():
						return
					case <-time.After(d):
						continue
					}
				}

				retries, start = healthCheckRetryTimes, time.Now()
				continue hcLoop
			}

			retries = 1 // fail fast
			if cerr := c.ReOpen(); cerr != nil {
				log.Errorf(cerr, "retry to health check etcd %s after %s", c.Endpoints, c.AutoSyncInterval)
			} else {
				log.Infof("[%s]re-connected to etcd %s", time.Now().Sub(start), c.Endpoints)
			}
		}
	}
}

func (c *EtcdClient) ReOpen() error {
	client, cerr := c.newClient()
	if cerr != nil {
		log.Errorf(cerr, "create a new connection to etcd %v failed",
			c.Endpoints)
		return cerr
	}
	c.Client, client = client, c.Client
	if cerr = client.Close(); cerr != nil {
		log.Errorf(cerr, "failed to close the unavailable etcd client")
	}
	client = nil
	return nil
}

func (c *EtcdClient) parseEndpoints() {
	// use the default cluster endpoints
	addrs := registry.Configuration().RegistryAddresses()

	endpoints := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if strings.Index(addr, "://") > 0 {
			// 如果配置格式为"sr-0=http(s)://IP:Port"，则需要分离IP:Port部分
			endpoints = append(endpoints, addr[strings.Index(addr, "://")+3:])
		} else {
			endpoints = append(endpoints, addr)
		}
	}

	c.Endpoints = endpoints
}

func (c *EtcdClient) SyncMembers(ctx context.Context) error {
	if err := c.Client.Sync(ctx); err != nil && err != c.Client.Ctx().Err() {
		return err
	}
	return nil
}

func dispatch(evts []*clientv3.Event, cb registry.WatchCallback) error {
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

func callback(action registry.ActionType, rev int64, kvs []*mvccpb.KeyValue, cb registry.WatchCallback) error {
	return cb("key information changed", &registry.PluginResponse{
		Action:    action,
		Kvs:       kvs,
		Count:     int64(len(kvs)),
		Revision:  rev,
		Succeeded: true,
	})
}

func NewRegistry() mgr.PluginInstance {
	log.Warnf("enable etcd registry mode")

	inst := &EtcdClient{}
	if err := inst.Initialize(); err != nil {
		inst.err <- err
		return inst
	}

	scheme := "http://"
	if inst.TLSConfig != nil {
		scheme = "https://"
	}
	firstEndpoint = scheme + inst.Endpoints[0]

	return inst
}
