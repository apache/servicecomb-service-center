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
package embededetcd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd/lease"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/ServiceComb/service-center/pkg/common"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/rest"
	"net/url"
	"strings"
	"time"
)

var embedTLSConfig *tls.Config

const START_MANAGER_SERVER_TIMEOUT = 60
const REGISTRY_PLUGIN_EMBEDED_ETCD = "embeded_etcd"

func init() {
	util.LOGGER.Warnf(nil, "embed etcd plugin init.")
	registry.RegistryPlugins[REGISTRY_PLUGIN_EMBEDED_ETCD] = getEmbedInstance
}

type EtcdEmbed struct {
	Server *embed.Etcd
	err    chan error
	ready  chan int
}

func (s *EtcdEmbed) Err() <-chan error {
	return s.err
}

func (s *EtcdEmbed) Ready() <-chan int {
	return s.ready
}

func (s *EtcdEmbed) Close() {
	if s.Server != nil {
		s.Server.Close()
	}
}

func (s *EtcdEmbed) toGetRequest(op *registry.PluginOp) *etcdserverpb.RangeRequest {
	endBytes := op.EndKey
	if op.WithPrefix {
		endBytes = append(op.Key, 127)
	}
	order := etcdserverpb.RangeRequest_NONE
	switch op.SortOrder {
	case registry.SORT_ASCEND:
		order = etcdserverpb.RangeRequest_ASCEND
	case registry.SORT_DESCEND:
		order = etcdserverpb.RangeRequest_DESCEND
	}
	return &etcdserverpb.RangeRequest{
		Key:        op.Key,
		RangeEnd:   endBytes,
		KeysOnly:   op.KeyOnly,
		CountOnly:  op.CountOnly,
		SortOrder:  order,
		SortTarget: etcdserverpb.RangeRequest_KEY,
		Revision:   op.WithRev,
	}
}

func (s *EtcdEmbed) toPutRequest(op *registry.PluginOp) *etcdserverpb.PutRequest {
	var valueBytes []byte
	if len(op.Value) > 0 {
		valueBytes = op.Value
	}
	return &etcdserverpb.PutRequest{
		Key:    op.Key,
		Value:  valueBytes,
		PrevKv: op.WithPrevKV,
		Lease:  op.Lease,
	}
}

func (s *EtcdEmbed) toDeleteRequest(op *registry.PluginOp) *etcdserverpb.DeleteRangeRequest {
	var endBytes []byte
	if op.WithPrefix {
		endBytes = append(op.Key, 127)
	}
	return &etcdserverpb.DeleteRangeRequest{
		Key:      op.Key,
		RangeEnd: endBytes,
		PrevKv:   op.WithPrevKV,
	}
}

func (s *EtcdEmbed) toTxnRequest(opts []*registry.PluginOp) []*etcdserverpb.RequestOp {
	etcdOps := []*etcdserverpb.RequestOp{}
	for _, op := range opts {
		switch op.Action {
		case registry.GET:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestRange{
					RequestRange: s.toGetRequest(op),
				},
			})
		case registry.PUT:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestPut{
					RequestPut: s.toPutRequest(op),
				},
			})
		case registry.DELETE:
			etcdOps = append(etcdOps, &etcdserverpb.RequestOp{
				Request: &etcdserverpb.RequestOp_RequestDeleteRange{
					RequestDeleteRange: s.toDeleteRequest(op),
				},
			})
		}
	}
	return etcdOps
}

func (s *EtcdEmbed) toCompares(cmps []*registry.CompareOp) []*etcdserverpb.Compare {
	etcdCmps := []*etcdserverpb.Compare{}
	for _, cmp := range cmps {
		compare := &etcdserverpb.Compare{
			Key: cmp.Key,
		}
		switch cmp.Type {
		case registry.CMP_VERSION:
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
		case registry.CMP_CREATE:
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
		case registry.CMP_MOD:
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
		case registry.CMP_VALUE:
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
		case registry.CMP_EQUAL:
			compare.Result = etcdserverpb.Compare_EQUAL
		case registry.CMP_GREATER:
			compare.Result = etcdserverpb.Compare_GREATER
		case registry.CMP_LESS:
			compare.Result = etcdserverpb.Compare_LESS
		case registry.CMP_NOT_EQUAL:
			compare.Result = etcdserverpb.Compare_NOT_EQUAL
		}
		etcdCmps = append(etcdCmps, compare)
	}
	return etcdCmps
}

func (s *EtcdEmbed) CompactCluster(ctx context.Context) {
}

func (s *EtcdEmbed) Compact(ctx context.Context, revision int64) error {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	revToCompact := max(0, revision-beego.AppConfig.DefaultInt64("compact_index_delta", 100))
	util.LOGGER.Debug(fmt.Sprintf("Compacting %d", revToCompact))
	resp, err := s.Server.Server.Compact(otCtx, &etcdserverpb.CompactionRequest{
		Revision: revToCompact,
	})
	if err != nil {
		return err
	}
	util.LOGGER.Info(fmt.Sprintf("Compacted %v", resp))
	return nil
}

func (s *EtcdEmbed) PutNoOverride(ctx context.Context, op *registry.PluginOp) (bool, error) {
	resp, err := s.TxnWithCmp(ctx, []*registry.PluginOp{op}, []*registry.CompareOp{
		{
			Key:    op.Key,
			Type:   registry.CMP_CREATE,
			Result: registry.CMP_EQUAL,
			Value:  0,
		},
	}, nil)
	util.LOGGER.Debugf("response %s %v %v", op.Key, resp.Succeeded, resp.Revision)
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}

func (s *EtcdEmbed) Do(ctx context.Context, op *registry.PluginOp) (*registry.PluginResponse, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	var err error
	var resp *registry.PluginResponse
	switch op.Action {
	case registry.GET:
		var etcdResp *etcdserverpb.RangeResponse
		etcdResp, err = s.Server.Server.Range(otCtx, s.toGetRequest(op))
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			Kvs:      etcdResp.Kvs,
			Count:    etcdResp.Count,
			Revision: etcdResp.Header.Revision,
		}
	case registry.PUT:
		var etcdResp *etcdserverpb.PutResponse
		etcdResp, err = s.Server.Server.Put(otCtx, s.toPutRequest(op))
		if err != nil {
			break
		}
		resp = &registry.PluginResponse{
			PrevKv:   etcdResp.PrevKv,
			Revision: etcdResp.Header.Revision,
		}
	case registry.DELETE:
		var etcdResp *etcdserverpb.DeleteRangeResponse
		etcdResp, err = s.Server.Server.DeleteRange(otCtx, s.toDeleteRequest(op))
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
func (s *EtcdEmbed) Txn(ctx context.Context, opts []*registry.PluginOp) (*registry.PluginResponse, error) {
	resp, err := s.TxnWithCmp(ctx, opts, nil, nil)
	if err != nil {
		return nil, err
	}
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Revision,
	}, nil
}

func (s *EtcdEmbed) TxnWithCmp(ctx context.Context, success []*registry.PluginOp, cmps []*registry.CompareOp, fail []*registry.PluginOp) (*registry.PluginResponse, error) {
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
	resp, err := s.Server.Server.Txn(otCtx, txnRequest)
	if err != nil {
		return nil, err
	}
	return &registry.PluginResponse{
		Succeeded: resp.Succeeded,
		Revision:  resp.Header.Revision,
	}, nil
}

func (s *EtcdEmbed) LeaseGrant(ctx context.Context, TTL int64) (int64, error) {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	etcdResp, err := s.Server.Server.LeaseGrant(otCtx, &etcdserverpb.LeaseGrantRequest{
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
	ttl, err := s.Server.Server.LeaseRenew(otCtx, lease.LeaseID(leaseID))
	if err != nil {
		return 0, err
	}
	return ttl, nil
}

func (s *EtcdEmbed) LeaseRevoke(ctx context.Context, leaseID int64) error {
	otCtx, cancel := registry.WithTimeout(ctx)
	defer cancel()
	_, err := s.Server.Server.LeaseRevoke(otCtx, &etcdserverpb.LeaseRevokeRequest{
		ID: leaseID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *EtcdEmbed) Watch(ctx context.Context, op *registry.PluginOp, send func(message string, evt *registry.PluginResponse) error) (err error) {
	if len(op.Key) > 0 {
		watchable := s.Server.Server.Watchable()
		ws := watchable.NewWatchStream()
		defer ws.Close()

		key := registry.BytesToStringWithNoCopy(op.Key)
		var keyBytes []byte
		if op.WithPrefix {
			key += "/"
			keyBytes = append([]byte(key), 127)
		}
		watchID := ws.Watch(op.Key, keyBytes, op.WithRev)
		// defer ws.Cancel(watchID)
		util.LOGGER.Infof("start to watch key %s, id is %d", key, watchID)

		responses := ws.Chan()
		for {
			select {
			case <-ctx.Done():
				util.LOGGER.Warnf(nil, "time out to watch key %s", key)
				return
			case resp, ok := <-responses:
				if !ok {
					util.LOGGER.Errorf(nil, "stop to watch key %s, id is %d, channel is closed", key, watchID)
					return
				}
				for _, evt := range resp.Events {
					pbEvent := &registry.PluginResponse{
						Action:    registry.PUT,
						Kvs:       []*mvccpb.KeyValue{evt.Kv},
						PrevKv:    evt.PrevKv,
						Count:     1,
						Revision:  resp.Revision,
						Succeeded: true,
					}
					if evt.Type == mvccpb.DELETE {
						pbEvent.Action = registry.DELETE
					}
					err = send("key information changed", pbEvent)
					if err != nil {
						util.LOGGER.Errorf(err, "stop to watch key %s, id is %d", key, watchID)
						return
					}
				}
			}
		}
	}
	err = fmt.Errorf("no key has been watched")
	util.LOGGER.Errorf(nil, err.Error())
	return
}

func getEmbedInstance(cfg *registry.Config) registry.Registry {
	util.LOGGER.Warnf(nil, "starting manager server in embed mode")

	hostName := beego.AppConfig.DefaultString("manager_name", util.GetLocalHostname())
	addrs := beego.AppConfig.String("manager_addr")

	inst := &EtcdEmbed{
		err:   make(chan error, 1),
		ready: make(chan int),
	}

	if common.GetServerSSLConfig().SSLEnabled {
		var err error
		embedTLSConfig, err = rest.GetServerTLSConfig(common.GetServerSSLConfig().VerifyClient)
		if err != nil {
			util.LOGGER.Error("get manager server tls config failed", err)
			inst.err <- err
			return inst
		}
	}

	serverCfg := embed.NewConfig()
	// TODO 不支持加密的TLS证书 ? managerTLSConfig
	// 存储目录，相对于工作目录
	serverCfg.Dir = "data"

	// 集群支持
	serverCfg.Name = hostName
	serverCfg.InitialCluster = cfg.ClusterAddresses

	// 管理端口
	urls, err := parseURL(addrs)
	if err != nil {
		util.LOGGER.Error(`"manager_addr" field configure error`, err)
		inst.err <- err
		return inst
	}
	serverCfg.LPUrls = urls
	serverCfg.APUrls = urls
	util.LOGGER.Debugf("--initial-cluster %s --initial-advertise-peer-urls %s --listen-peer-urls %s",
		serverCfg.InitialCluster, addrs, addrs)

	// 业务端口，关闭默认2379端口
	// clients := beego.AppConfig.String("clientcluster")
	serverCfg.LCUrls = nil
	serverCfg.ACUrls = nil

	// 自动压缩历史, 1 hour
	serverCfg.AutoCompactionRetention = 1

	etcd, err := embed.StartEtcd(serverCfg)
	if err != nil {
		util.LOGGER.Error("error to start etcd server", err)
		inst.err <- err
		return inst
	}
	inst.Server = etcd

	select {
	case <-etcd.Server.ReadyNotify():
		close(inst.ready)
		go func() {
			select {
			case err := <-etcd.Err():
				inst.err <- err
			}
		}()
	case <-time.After(START_MANAGER_SERVER_TIMEOUT * time.Second):
		message := "etcd server took too long to start"
		util.LOGGER.Error(message, nil)

		etcd.Server.Stop()

		inst.err <- errors.New(message)
	}
	return inst
}

func parseURL(addrs string) ([]url.URL, error) {
	urls := []url.URL{}
	ips := strings.Split(addrs, ",")
	for _, ip := range ips {
		addr, err := url.Parse(ip)
		if err != nil {
			util.LOGGER.Error("Error to parse ip address string", err)
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
