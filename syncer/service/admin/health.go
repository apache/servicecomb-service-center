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
	"time"

	"github.com/apache/servicecomb-service-center/client"
	grpc "github.com/apache/servicecomb-service-center/pkg/rpc"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	syncerclient "github.com/apache/servicecomb-service-center/syncer/client"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/metrics"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
)

const (
	scheme      = "health_rpc"
	serviceName = "syncer"
)

var (
	ErrConfigIsEmpty = errors.New("sync config is empty")
)

type Resp struct {
	Peers []*Peer `json:"peers"`
}

type Peer struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Mode      []string `json:"mode"`
	Endpoints []string `json:"endpoints"`
	Status    string   `json:"status"`
}

func Health() (*Resp, error) {
	cfg := config.GetConfig()
	if cfg.Sync == nil || len(cfg.Sync.Peers) <= 0 {
		return nil, ErrConfigIsEmpty
	}

	resp := &Resp{Peers: make([]*Peer, 0, len(cfg.Sync.Peers))}

	for _, c := range cfg.Sync.Peers {
		if len(c.Endpoints) <= 0 {
			continue
		}
		p := &Peer{
			Name:      c.Name,
			Kind:      c.Kind,
			Mode:      c.Mode,
			Endpoints: c.Endpoints,
		}
		p.Status = getPeerStatus(c.Name, c.Endpoints)
		resp.Peers = append(resp.Peers, p)
	}

	reportMetrics(resp.Peers)

	if len(resp.Peers) <= 0 {
		return nil, ErrConfigIsEmpty
	}

	return resp, nil
}

func getPeerStatus(peerName string, endpoints []string) string {
	conn, err := grpc.GetRoundRobinLbConn(&grpc.Config{
		Addrs:       endpoints,
		Scheme:      scheme,
		ServiceName: serviceName,
		TLSConfig:   syncerclient.RPClientConfig(),
	})
	if err != nil || conn == nil {
		return rpc.HealthStatusAbnormal
	}
	defer conn.Close()

	local := time.Now().UnixNano()
	set := client.NewSet(conn)
	reply, err := set.EventServiceClient.Health(context.Background(), &v1sync.HealthRequest{})
	if err != nil || reply == nil {
		return rpc.HealthStatusAbnormal
	}
	reportClockDiff(peerName, local, reply.LocalTimestamp)
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

	metrics.ConnectedPeersSet(int64(len(peers)))
	metrics.PeersTotalSet(connectPeersCount)
}
