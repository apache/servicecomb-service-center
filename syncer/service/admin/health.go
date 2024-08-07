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

package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-chassis/go-chassis/v2/server/restful"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pkgrpc "github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/server/plugin/security/cipher"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	syncerclient "github.com/apache/servicecomb-service-center/syncer/client"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/metrics"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
)

const (
	scheme      = "grpc"
	serviceName = "syncer"
)

var (
	peerInfos        []*PeerInfo
	ErrConfigIsEmpty = errors.New("sync config is empty")
)

type Resp struct {
	Peers []*Peer `json:"peers"`
}

type PeerInfo struct {
	Peer       *Peer
	ClientConn *grpc.ClientConn
}

type Peer struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Mode      []string `json:"mode"`
	Endpoints []string `json:"endpoints"`
	Status    string   `json:"status"`
	Token     string   `json:"-"`
}

func Init() {
	cfg := config.GetConfig()
	peerInfos = make([]*PeerInfo, 0, len(cfg.Sync.Peers))
	for _, c := range cfg.Sync.Peers {
		if len(c.Endpoints) <= 0 {
			log.Warn("no endpoints of peer: " + c.Name)
			continue
		}
		p := &Peer{
			Name:      c.Name,
			Kind:      c.Kind,
			Mode:      c.Mode,
			Endpoints: c.Endpoints,
		}
		if config.GetConfig().Sync.RbacEnabled {
			plainToken, err := cipher.Decrypt(c.Token)
			if err != nil {
				log.Error(fmt.Sprintf("decrypt token of peer %s failed, use original content", c.Name), err)
				plainToken = c.Token
			}
			p.Token = plainToken
		}

		conn, err := newRPCConn(p.Endpoints)
		if err != nil {
			log.Error(fmt.Sprintf("new client failed for peer: %s", c.Name), err)
			continue
		}
		peerInfos = append(peerInfos, &PeerInfo{Peer: p, ClientConn: conn})
	}
}

func Health() (*Resp, error) {
	if len(peerInfos) <= 0 {
		return nil, ErrConfigIsEmpty
	}

	resp := &Resp{Peers: make([]*Peer, 0, len(peerInfos))}
	for _, peerInfo := range peerInfos {
		if len(peerInfo.Peer.Endpoints) <= 0 {
			continue
		}
		status := getPeerStatus(peerInfo)
		resp.Peers = append(resp.Peers, &Peer{
			Name:      peerInfo.Peer.Name,
			Kind:      peerInfo.Peer.Kind,
			Mode:      peerInfo.Peer.Mode,
			Endpoints: peerInfo.Peer.Endpoints,
			Status:    status,
		})
	}

	reportMetrics(resp.Peers)
	return resp, nil
}

func getPeerStatus(peerInfo *PeerInfo) string {
	if peerInfo.ClientConn == nil {
		log.Warn("clientConn is nil")
		return rpc.HealthStatusAbnormal
	}
	local := time.Now().UnixNano()
	set := client.NewSet(peerInfo.ClientConn)
	ctx := context.Background()
	if config.GetConfig().Sync.RbacEnabled {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			restful.HeaderAuth: "Bearer " + peerInfo.Peer.Token,
		}))
	}
	reply, err := set.EventServiceClient.Health(ctx, &v1sync.HealthRequest{})
	if err != nil || reply == nil {
		log.Error("get peer health failed", err)
		return rpc.HealthStatusAbnormal
	}
	reportClockDiff(peerInfo.Peer.Name, local, reply.LocalTimestamp)
	return reply.Status
}

func reportClockDiff(peerName string, local int64, resp int64) {
	curr := time.Now().UnixNano()
	spent := (curr - local) / 2
	metrics.PeersClockDiffSet(peerName, local+spent-resp)
}

func reportMetrics(peers []*Peer) {
	var connectPeersCount int64
	for _, peer := range peers {
		if peer.Status == rpc.HealthStatusConnected {
			connectPeersCount++
		}
	}

	metrics.PeersTotalSet(int64(len(peers)))
	metrics.ConnectedPeersSet(connectPeersCount)
}

func newRPCConn(endpoints []string) (*grpc.ClientConn, error) {
	return pkgrpc.GetRoundRobinLbConn(&pkgrpc.Config{
		Addrs:       endpoints,
		Scheme:      scheme,
		ServiceName: serviceName,
		TLSConfig:   syncerclient.RPClientConfig(),
	})
}

func Peers() []*PeerInfo {
	return peerInfos
}
