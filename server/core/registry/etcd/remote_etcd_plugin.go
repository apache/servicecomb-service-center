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
package etcd

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/ServiceComb/service-center/pkg/common"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/rest"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"golang.org/x/net/context"
	"strings"
	"time"
)

const (
	REGISTRY_PLUGIN_ETCD           = "etcd"
	CONNECT_MANAGER_SERVER_TIMEOUT = 10
)

var clientTLSConfig *tls.Config

func init() {
	util.LOGGER.Warnf(nil, "etcd plugin init.")
	registry.RegistryPlugins[REGISTRY_PLUGIN_ETCD] = NewRegistry
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
	util.LOGGER.Debugf("etcd client stopped.")
}

func (c *EtcdClient) CompactCluster(ctx context.Context) {
	for _, ep := range c.Client.Endpoints() {
		otCtx, cancel := registry.WithTimeout(ctx)
		defer cancel()
		mapi := clientv3.NewMaintenance(c.Client)
		resp, err := mapi.Status(otCtx, ep)
		if err != nil {
			util.LOGGER.Error(fmt.Sprintf("Compact error ,can not get status from %s", ep), err)
			continue
		}
		curRev := resp.Header.Revision
		util.LOGGER.Debug(fmt.Sprintf("Compacting.... endpoint: %s / IsLeader: %v\n / revision is %d", ep, resp.Header.MemberId == resp.Leader, curRev))
		c.Compact(ctx, curRev)
	}

}

func (c *EtcdClient) Compact(ctx context.Context, revision int64) error {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	revToCompact := max(0, revision-beego.AppConfig.DefaultInt64("compact_index_delta", 100))
	if revToCompact <= 0 {
		util.LOGGER.Warnf(nil, "revToCompact is %d, <=0, no nead to compact.", revToCompact)
		return nil
	}
	util.LOGGER.Debug(fmt.Sprintf("Compacting %d", revToCompact))
	resp, err := c.Client.KV.Compact(otCtx, revToCompact)
	if err != nil {
		return err
	}
	util.LOGGER.Debugf(fmt.Sprintf("Compacted %v", resp))
	return nil
}

func max(n1, n2 int64) int64 {
	if n1 > n2 {
		return n1
	}
	return n2
}

func (s *EtcdClient) toGetRequest(op *registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	} else if len(op.EndKey) > 0 {
		opts = append(opts, clientv3.WithRange(registry.BytesToStringWithNoCopy(op.EndKey)))
	}
	if op.WithPrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	if op.KeyOnly {
		opts = append(opts, clientv3.WithKeysOnly())
	}
	if op.CountOnly {
		opts = append(opts, clientv3.WithCountOnly())
	}
	if op.WithRev > 0 {
		opts = append(opts, clientv3.WithRev(op.WithRev))
	}
	switch op.SortOrder {
	case registry.SORT_ASCEND:
		opts = append(opts, clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	case registry.SORT_DESCEND:
		opts = append(opts, clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	}
	return opts
}

func (s *EtcdClient) toPutRequest(op *registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.WithPrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	if op.Lease > 0 {
		opts = append(opts, clientv3.WithLease(clientv3.LeaseID(op.Lease)))
	}
	return opts
}

func (s *EtcdClient) toDeleteRequest(op *registry.PluginOp) []clientv3.OpOption {
	opts := []clientv3.OpOption{}
	if op.WithPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	if op.WithPrevKV {
		opts = append(opts, clientv3.WithPrevKV())
	}
	return opts
}

func (c *EtcdClient) toTxnRequest(opts []*registry.PluginOp) []clientv3.Op {
	etcdOps := []clientv3.Op{}
	for _, op := range opts {
		switch op.Action {
		case registry.GET:
			etcdOps = append(etcdOps, clientv3.OpGet(registry.BytesToStringWithNoCopy(op.Key), c.toGetRequest(op)...))
		case registry.PUT:
			var value string = ""
			if len(op.Value) > 0 {
				value = registry.BytesToStringWithNoCopy(op.Value)
			}
			etcdOps = append(etcdOps, clientv3.OpPut(registry.BytesToStringWithNoCopy(op.Key), value, c.toPutRequest(op)...))
		case registry.DELETE:
			etcdOps = append(etcdOps, clientv3.OpDelete(registry.BytesToStringWithNoCopy(op.Key), c.toDeleteRequest(op)...))
		}
	}
	return etcdOps
}

func (c *EtcdClient) toCompares(cmps []*registry.CompareOp) []clientv3.Cmp {
	etcdCmps := []clientv3.Cmp{}
	for _, cmp := range cmps {
		var cmpType clientv3.Cmp
		var cmpResult string
		key := registry.BytesToStringWithNoCopy(cmp.Key)
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

func (c *EtcdClient) PutNoOverride(ctx context.Context, op *registry.PluginOp) (bool, error) {
	resp, err := c.TxnWithCmp(ctx, []*registry.PluginOp{op}, []*registry.CompareOp{
		{
			Key:    op.Key,
			Type:   registry.CMP_CREATE,
			Result: registry.CMP_EQUAL,
			Value:  0,
		},
	}, nil)
	if err != nil {
		util.LOGGER.Errorf(err, "PutNoOverride %s failed", op.Key)
		return false, err
	}
	util.LOGGER.Infof("response %s %v %v", op.Key, resp.Succeeded, resp.Revision)
	return resp.Succeeded, nil
}

func (c *EtcdClient) Do(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	var err error
	var resp *registry.PluginResponse
	switch op.Action {
	case registry.GET:
		var etcdResp *clientv3.GetResponse
		etcdResp, err = c.Client.Get(otCtx, registry.BytesToStringWithNoCopy(op.Key), c.toGetRequest(op)...)
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Kvs:      etcdResp.Kvs,
			Count:    etcdResp.Count,
			Revision: etcdResp.Header.Revision,
		}
	case registry.PUT:
		var value string = ""
		if len(op.Value) > 0 {
			value = registry.BytesToStringWithNoCopy(op.Value)
		}
		var etcdResp *clientv3.PutResponse
		etcdResp, err = c.Client.Put(otCtx, registry.BytesToStringWithNoCopy(op.Key), value, c.toPutRequest(op)...)
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			PrevKv:   etcdResp.PrevKv,
			Revision: etcdResp.Header.Revision,
		}
	case registry.DELETE:
		var etcdResp *clientv3.DeleteResponse
		etcdResp, err = c.Client.Delete(otCtx, registry.BytesToStringWithNoCopy(op.Key), c.toDeleteRequest(op)...)
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

func (c *EtcdClient) Txn(ctx context.Context, opts []*registry.PluginOp) (*registry.PluginResponse, error) {
	resp, err := c.TxnWithCmp(ctx, opts, nil, nil)
	if err != nil {
		return nil, err
	}
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Revision,
	}, nil
}

func (c *EtcdClient) TxnWithCmp(ctx context.Context, success []*registry.PluginOp, cmps []*registry.CompareOp, fail []*registry.PluginOp) (*registry.PluginResponse, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()

	etcdCmps := c.toCompares(cmps)
	etcdSuccessOps := c.toTxnRequest(success)
	etcdFailOps := c.toTxnRequest(fail)

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
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Header.Revision,
	}, nil
}

func (c *EtcdClient) LeaseGrant(ctx context.Context, TTL int64) (int64, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	etcdResp, err := c.Client.Grant(otCtx, TTL)
	if err != nil {
		return 0, err
	}
	return int64(etcdResp.ID), nil
}

func (c *EtcdClient) LeaseRenew(ctx context.Context, leaseID int64) (int64, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	etcdResp, err := c.Client.KeepAliveOnce(otCtx, clientv3.LeaseID(leaseID))
	if err != nil {
		return 0, err
	}
	return etcdResp.TTL, nil
}

func (c *EtcdClient) LeaseRevoke(ctx context.Context, leaseID int64) error {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	_, err := c.Client.Revoke(otCtx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}
	return nil
}

func (c *EtcdClient) Watch(ctx context.Context, op *registry.PluginOp, send func(message string, evt *registry.PluginResponse) error) (err error) {
	n := len(op.Key)
	if n > 0 {
		// 必须创建新的client连接
		/*client, err := newClient(c.Client.Endpoints())
		  if err != nil {
		          util.LOGGER.Error("get manager client failed", err)
		          return err
		  }
		  defer client.Close()*/
		// 现在跟ETCD仅使用一个连接，共用client即可
		client := clientv3.NewWatcher(c.Client)
		defer client.Close()

		key := registry.BytesToStringWithNoCopy(op.Key)
		if op.WithPrefix {
			key += "/"
		}
		util.LOGGER.Debugf("start to watch key %s", key)

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
					err := errors.New("channel is closed")
					return err
				}
				if err = resp.Err(); err != nil {
					return err
				}
				for _, evt := range resp.Events {
					pbEvent := &registry.PluginResponse{
						Action:    registry.PUT,
						Kvs:       []*mvccpb.KeyValue{evt.Kv},
						PrevKv:    evt.PrevKv,
						Count:     1,
						Revision:  evt.Kv.ModRevision,
						Succeeded: true,
					}
					if evt.Type == mvccpb.DELETE {
						pbEvent.Action = registry.DELETE
					}

					err = send("key information changed", pbEvent)
					if err != nil {
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("no key has been watched")
	return
}

func NewRegistry(cfg *registry.Config) registry.Registry {
	util.LOGGER.Warnf(nil, "starting manager server in proxy mode")

	inst := &EtcdClient{
		err:   make(chan error, 1),
		ready: make(chan int),
	}
	addrs := strings.Split(cfg.ClusterAddresses, ",")

	if common.GetClientSSLConfig().SSLEnabled && strings.Index(cfg.ClusterAddresses, "https://") >= 0 {
		var err error
		// go client tls限制，提供身份证书、不认证服务端、不校验CN
		clientTLSConfig, err = rest.GetClientTLSConfig(common.GetClientSSLConfig().VerifyClient, true, false)
		if err != nil {
			util.LOGGER.Error("get etcd client tls config failed", err)
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
	refreshManagerClusterInterval := cfg.AutoSyncInterval
	util.LOGGER.Debugf("refreshManagerClusterInterval is %d", refreshManagerClusterInterval)
	client, err := newClient(endpoints, refreshManagerClusterInterval)
	if err != nil {
		util.LOGGER.Errorf(err, "get etcd client %+v failed.", endpoints)
		inst.err <- err
		return inst
	}

	util.LOGGER.Warnf(nil, "get etcd client %+v completed.", endpoints)
	inst.Client = client
	close(inst.ready)
	return inst
}

func newClient(endpoints []string, autoSyncInterval int64) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:        endpoints,
		DialTimeout:      CONNECT_MANAGER_SERVER_TIMEOUT * time.Second,
		TLS:              clientTLSConfig, // 暂时与API Server共用一套证书
		AutoSyncInterval: time.Duration(autoSyncInterval) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
