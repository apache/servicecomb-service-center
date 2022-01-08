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

	v1sync "github.com/apache/servicecomb-service-center/api/sync/v1"
	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/server/rpc/sync"
	"github.com/apache/servicecomb-service-center/syncer/config"
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

	config := config.GetConfig()
	if config.Sync == nil || len(config.Sync.Peers) <= 0 {
		return nil, ErrConfigIsEmpty
	}

	resp := &Resp{Peers: make([]*Peer, 0, len(config.Sync.Peers))}

	for _, c := range config.Sync.Peers {
		if len(c.Endpoints) <= 0 {
			continue
		}
		p := &Peer{
			Name:      c.Name,
			Kind:      c.Kind,
			Mode:      c.Mode,
			Endpoints: c.Endpoints,
		}
		p.Status = getPeerStatus(c.Endpoints)
		resp.Peers = append(resp.Peers, p)
	}

	if len(resp.Peers) <= 0 {
		return nil, ErrConfigIsEmpty
	}

	return resp, nil
}

func getPeerStatus(endpoints []string) string {
	conn, err := rpc.GetRoundRobinLbConn(&rpc.Config{Addrs: endpoints, Scheme: scheme, ServiceName: serviceName})
	if err != nil || conn == nil {
		return sync.HealthStatusAbnormal
	}
	defer conn.Close()

	set := client.NewSet(conn)
	reply, err := set.EventServiceClient.Health(context.Background(), &v1sync.HealthRequest{})
	if err != nil || reply == nil {
		return sync.HealthStatusAbnormal
	}
	return reply.Status
}
