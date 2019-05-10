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
package peer

import (
	"context"
	"io"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"os"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

// Agent warps the serf agent
type Agent struct {
	*agent.Agent
	conf *Config
}

// Create create peer agent with config
func Create(peerConf *Config, logOutput io.Writer) (*Agent, error) {
	if logOutput == nil {
		logOutput = os.Stderr
	}

	// peer config cover to serf config
	serfConf, err := peerConf.convertToSerf()
	if err != nil {
		return nil, err
	}

	// create serf agent with serf config
	serfAgent, err := agent.Create(peerConf.Config, serfConf, logOutput)
	if err != nil {
		return nil, err
	}
	return &Agent{Agent: serfAgent, conf: peerConf}, nil
}

// Start agent
func (a *Agent) Start(ctx context.Context) {
	err := a.Agent.Start()
	if err != nil {
		log.Errorf(err, "start peer agent failed")
	}

	gopool.Go(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			a.Leave()
			a.Shutdown()
		}
	})
}

// Leave from Serf
func (a *Agent) Leave() error {
	return a.Agent.Leave()
}

// Shutdown Serf server
func (a *Agent) Shutdown() error {
	return a.Agent.Shutdown()
}

// ShutdownCh returns a channel that can be selected to wait
// for the agent to perform a shutdown.
func (a *Agent) ShutdownCh() <-chan struct{} {
	return a.Agent.ShutdownCh()
}

func (a *Agent) LocalMember() *serf.Member {
	serfAgent := a.Agent.Serf()
	if serfAgent != nil {
		member := serfAgent.LocalMember()
		return &member
	}
	serfAgent.State()
	return nil
}

// Member get member information with node
func (a *Agent) Member(node string) *serf.Member {
	serfAgent := a.Agent.Serf()
	if serfAgent != nil {
		ms := serfAgent.Members()
		for _, m := range ms {
			if m.Name == node {
				return &m
			}
		}
	}
	return nil
}

// SerfConfig get serf config
func (a *Agent) SerfConfig() *serf.Config {
	return a.Agent.SerfConfig()
}

// Join serf clusters through one or more members
func (a *Agent) Join(addrs []string, replay bool) (n int, err error) {
	return a.Agent.Join(addrs, replay)
}

// ForceLeave forced to leave from serf
func (a *Agent) ForceLeave(node string) error {
	return a.Agent.ForceLeave(node)
}

// UserEvent sends a UserEvent on Serf
func (a *Agent) UserEvent(name string, payload []byte, coalesce bool) error {
	return a.Agent.UserEvent(name, payload, coalesce)
}

// Query sends a Query on Serf
func (a *Agent) Query(name string, payload []byte, params *serf.QueryParam) (*serf.QueryResponse, error) {
	return a.Agent.Query(name, payload, params)
}
