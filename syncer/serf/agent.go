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
package serf

import (
	"context"
	"errors"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

// Agent warps the serf agent
type Agent struct {
	*agent.Agent
	conf    *Config
	readyCh chan struct{}
	errorCh chan error
}

// Create create serf agent with config
func Create(conf *Config) (*Agent, error) {
	// config cover to serf config
	serfConf, err := conf.convertToSerf()
	if err != nil {
		return nil, err
	}

	// create serf agent with serf config
	serfAgent, err := agent.Create(conf.Config, serfConf, nil)
	if err != nil {
		return nil, err
	}
	return &Agent{
		Agent:   serfAgent,
		conf:    conf,
		readyCh: make(chan struct{}),
		errorCh: make(chan error),
	}, nil
}

// Start agent
func (a *Agent) Start(ctx context.Context) {
	err := a.Agent.Start()
	if err != nil {
		log.Errorf(err, "start serf agent failed")
		a.errorCh <- err
		return
	}
	a.RegisterEventHandler(a)

	err = a.retryJoin(ctx)
	if err != nil {
		log.Errorf(err, "start serf agent failed")
		if err != ctx.Err() && a.errorCh != nil {
			a.errorCh <- err
		}
	}
}

// HandleEvent Handles serf.EventMemberJoin events,
// which will wait for members to join until the number of group members is equal to "groupExpect"
// when the startup mode is "ModeCluster",
// used for logical grouping of serf nodes
func (a *Agent) HandleEvent(event serf.Event) {
	if event.EventType() != serf.EventMemberJoin {
		return
	}

	if a.conf.Mode == ModeCluster {
		if len(a.GroupMembers(a.conf.ClusterName)) < groupExpect {
			return
		}
	}
	a.DeregisterEventHandler(a)
	close(a.readyCh)
}

// Ready Returns a channel that will be closed when serf is ready
func (a *Agent) Ready() <-chan struct{} {
	return a.readyCh
}

// Error Returns a channel that will be transmit a serf error
func (a *Agent) Error() <-chan error {
	return a.errorCh
}

// Stop serf agent
func (a *Agent) Stop() {
	if a.errorCh != nil {
		a.Leave()
		a.Shutdown()
		close(a.errorCh)
		a.errorCh = nil
	}
}

// LocalMember returns the Member information for the local node
func (a *Agent) LocalMember() *serf.Member {
	serfAgent := a.Agent.Serf()
	if serfAgent != nil {
		member := serfAgent.LocalMember()
		return &member
	}
	return nil
}

// GroupMembers returns a point-in-time snapshot of the members of by groupName
func (a *Agent) GroupMembers(groupName string) (members []serf.Member) {
	serfAgent := a.Agent.Serf()
	if serfAgent != nil {
		for _, member := range serfAgent.Members() {
			log.Infof("member = %s, groupName = %s", member.Name, member.Tags[tagKeyCluster])
			if member.Tags[tagKeyCluster] == a.conf.ClusterName {
				members = append(members, member)
			}
		}
	}
	return
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

// UserEvent sends a UserEvent on Serf
func (a *Agent) UserEvent(name string, payload []byte, coalesce bool) error {
	return a.Agent.UserEvent(name, payload, coalesce)
}

// Query sends a Query on Serf
func (a *Agent) Query(name string, payload []byte, params *serf.QueryParam) (*serf.QueryResponse, error) {
	return a.Agent.Query(name, payload, params)
}

func (a *Agent) retryJoin(ctx context.Context) (err error) {
	if len(a.conf.RetryJoin) == 0 {
		log.Infof("retry join mumber %d", len(a.conf.RetryJoin))
		return nil
	}

	// Count of attempts
	attempt := 0
	ticker := time.NewTicker(a.conf.RetryInterval)
	for {
		log.Infof("serf: Joining cluster...(replay: %v)", a.conf.ReplayOnJoin)
		var n int

		// Try to join the specified serf nodes
		n, err = a.Join(a.conf.RetryJoin, a.conf.ReplayOnJoin)
		if err == nil {
			log.Infof("serf: Join completed. Synced with %d initial agents", n)
			break
		}
		attempt++

		// If RetryMaxAttempts is greater than 0, agent will exit
		// and throw an error when the number of attempts exceeds RetryMaxAttempts,
		// else agent will try to join other nodes until successful always
		if a.conf.RetryMaxAttempts > 0 && attempt > a.conf.RetryMaxAttempts {
			err = errors.New("serf: maximum retry join attempts made, exiting")
			log.Errorf(err, err.Error())
			break
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			goto done
		// Waiting for ticker to trigger
		case <-ticker.C:
		}
	}
done:
	ticker.Stop()
	return
}
