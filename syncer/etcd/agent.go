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
	"context"
	"errors"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/etcdserver/api/v3client"
)

// Agent warps the embed etcd
type Agent struct {
	conf    *Config
	etcd    *embed.Etcd
	readyCh chan struct{}
	stopCh  chan struct{}
}

// NewAgent new etcd agent with config
func NewAgent(conf *Config) *Agent {
	return &Agent{
		conf:    conf,
		readyCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}
}

// Start etcd agent
func (a *Agent) Start(ctx context.Context) {
	etcd, err := embed.StartEtcd(a.conf.Config)
	if err == nil {
		a.etcd = etcd
		select {
		// Be returns when the server is readied
		case <-etcd.Server.ReadyNotify():
			log.Info("start etcd success")
			close(a.readyCh)

		// Be returns when the server is stopped
		case <-etcd.Server.StopNotify():
			err = errors.New("unknown error cause start etcd failed, check etcd")

		// Returns an error when running goroutine fails in the etcd startup process
		case err = <-etcd.Err():
		case <-ctx.Done():
			err = errors.New("cancel etcd server from context")
		}
	}

	if err != nil {
		log.Error("start etcd failed", err)
		close(a.stopCh)
	}
}

// Ready Returns a channel that will be closed when etcd is ready
func (a *Agent) Ready() <-chan struct{} {
	return a.readyCh
}

// Error Returns a channel that will be closed when etcd is stopped
func (a *Agent) Stopped() <-chan struct{} {
	return a.stopCh
}

// Storage returns etcd storage
func (a *Agent) Storage() *clientv3.Client {
	return v3client.New(a.etcd.Server)
}

// Stop etcd agent
func (a *Agent) Stop() {
	if a.etcd != nil {
		a.etcd.Close()
	}
}

// IsLeader Check leader
func (a *Agent) IsLeader() bool {
	if a.etcd == nil || a.etcd.Server == nil {
		return false
	}
	return a.etcd.Server.Leader() == a.etcd.Server.ID()
}
