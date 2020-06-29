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
package etcdsync

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/coreos/etcd/client"
	"os"
	"sync"
	"time"
)

const (
	DEFAULT_LOCK_TTL    = 60
	DEFAULT_RETRY_TIMES = 3
	ROOT_PATH           = "/cse/etcdsync"

	OperationGlobalLock = "GLOBAL_LOCK"
)

type DLock struct {
	key      string
	ctx      context.Context
	ttl      int64
	mutex    *sync.Mutex
	id       string
	createAt time.Time
}

var (
	globalMux sync.Mutex
	IsDebug   bool
	hostname  = util.HostName()
	pid       = os.Getpid()
	mutex     = new(sync.Mutex)
)

func NewDLock(key string, ttl int64, wait bool) (l *DLock, err error) {
	if len(key) == 0 {
		return nil, nil
	}
	if ttl < 1 {
		ttl = DEFAULT_LOCK_TTL
	}

	now := time.Now()
	l = &DLock{
		key:      key,
		ctx:      context.Background(),
		ttl:      ttl,
		id:       fmt.Sprintf("%v-%v-%v", hostname, pid, now.Format("20060102-15:04:05.999999999")),
		createAt: now,
		mutex:    &sync.Mutex{},
	}
	for try := 1; try <= DEFAULT_RETRY_TIMES; try++ {
		err = l.Lock(wait)
		if err == nil {
			return
		}

		if !wait {
			break
		}
	}
	// failed
	log.Errorf(err, "Lock key %s failed, id=%s", l.key, l.id)
	l = nil
	return
}

func (m *DLock) ID() string {
	return m.id
}

func (m *DLock) Lock(wait bool) (err error) {
	if !IsDebug {
		m.mutex.Lock()
	}

	opts := []registry.PluginOpOption{
		registry.WithStrKey(m.key),
		registry.WithStrValue(m.id)}

	log.Infof("Trying to create a lock: key=%s, id=%s", m.key, m.id)

	var leaseID int64
	putOpts := opts
	if m.ttl > 0 {
		leaseID, err = backend.Registry().LeaseGrant(m.ctx, m.ttl)
		if err != nil {
			return err
		}
		putOpts = append(opts, registry.WithLease(leaseID))
	}
	success, err := backend.Registry().PutNoOverride(m.ctx, putOpts...)
	if err == nil && success {
		log.Infof("Create Lock OK, key=%s, id=%s", m.key, m.id)
		return nil
	}

	if leaseID > 0 {
		backend.Registry().LeaseRevoke(m.ctx, leaseID)
	}

	if m.ttl == 0 || !wait {
		return fmt.Errorf("key %s is locked by id=%s", m.key, m.id)
	}

	log.Errorf(err, "Key %s is locked, waiting for other node releases it, id=%s", m.key, m.id)

	ctx, cancel := context.WithTimeout(m.ctx, time.Duration(m.ttl)*time.Second)
	gopool.Go(func(context.Context) {
		defer cancel()
		err := backend.Registry().Watch(ctx,
			registry.WithStrKey(m.key),
			registry.WithWatchCallback(
				func(message string, evt *registry.PluginResponse) error {
					if evt != nil && evt.Action == registry.Delete {
						// break this for-loop, and try to create the node again.
						return fmt.Errorf("lock released")
					}
					return nil
				}))
		if err != nil {
			log.Warnf("%s, key=%s, id=%s", err.Error(), m.key, m.id)
		}
	})
	select {
	case <-ctx.Done():
		return ctx.Err() // 可以重新尝试获取锁
	case <-m.ctx.Done():
		cancel()
		return m.ctx.Err() // 机制错误，不应该超时的
	}
}

func (m *DLock) Unlock() (err error) {
	defer func() {
		if !IsDebug {
			m.mutex.Unlock()
		}

		registry.ReportBackendOperationCompleted(OperationGlobalLock, nil, m.createAt)
	}()

	opts := []registry.PluginOpOption{
		registry.DEL,
		registry.WithStrKey(m.key)}

	for i := 1; i <= DEFAULT_RETRY_TIMES; i++ {
		_, err = backend.Registry().Do(m.ctx, opts...)
		if err == nil {
			log.Infof("Delete lock OK, key=%s, id=%s", m.key, m.id)
			return nil
		}
		log.Errorf(err, "Delete lock failed, key=%s, id=%s", m.key, m.id)
		e, ok := err.(client.Error)
		if ok && e.Code == client.ErrorCodeKeyNotFound {
			return nil
		}
	}
	return err
}

func Lock(key string, ttl int64, wait bool) (*DLock, error) {
	return NewDLock(fmt.Sprintf("%s%s", ROOT_PATH, key), ttl, wait)
}
