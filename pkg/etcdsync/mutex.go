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
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"os"
	"sync"
	"time"
)

const (
	DEFAULT_LOCK_TTL    = 60
	DEFAULT_RETRY_TIMES = 3
	ROOT_PATH           = "/cse/etcdsync"
)

type DLockFactory struct {
	key   string
	ctx   context.Context
	ttl   int64
	mutex *sync.Mutex
}

type DLock struct {
	builder *DLockFactory
	id      string
}

var (
	globalMap = make(map[string]*DLockFactory)
	globalMux sync.Mutex
	IsDebug   bool
	hostname  = util.HostName()
	pid       = os.Getpid()
)

// lock will not be release automatically if ttl = 0
func NewLockFactory(key string, ttl int64) *DLockFactory {
	if len(key) == 0 {
		return nil
	}
	if ttl < 1 {
		ttl = DEFAULT_LOCK_TTL
	}

	return &DLockFactory{
		key:   key,
		ctx:   context.Background(),
		ttl:   ttl,
		mutex: new(sync.Mutex),
	}
}

func (m *DLockFactory) NewDLock(wait bool) (l *DLock, err error) {
	if !IsDebug {
		m.mutex.Lock()
	}
	l = &DLock{
		builder: m,
		id:      fmt.Sprintf("%v-%v-%v", hostname, pid, time.Now().Format("20060102-15:04:05.999999999")),
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
	log.Errorf(err, "Lock key %s failed, id=%s", m.key, l.id)
	l = nil

	if !IsDebug {
		m.mutex.Unlock()
	}
	return
}

func (m *DLock) ID() string {
	return m.id
}

func (m *DLock) Lock(wait bool) (err error) {
	opts := []registry.PluginOpOption{
		registry.WithStrKey(m.builder.key),
		registry.WithStrValue(m.id)}

	log.Infof("Trying to create a lock: key=%s, id=%s", m.builder.key, m.id)

	var leaseID int64
	putOpts := opts
	if m.builder.ttl > 0 {
		leaseID, err = backend.Registry().LeaseGrant(m.builder.ctx, m.builder.ttl)
		if err != nil {
			return err
		}
		putOpts = append(opts, registry.WithLease(leaseID))
	}
	success, err := backend.Registry().PutNoOverride(m.builder.ctx, putOpts...)
	if err == nil && success {
		log.Infof("Create Lock OK, key=%s, id=%s", m.builder.key, m.id)
		return nil
	}

	if leaseID > 0 {
		backend.Registry().LeaseRevoke(m.builder.ctx, leaseID)
	}

	if m.builder.ttl == 0 || !wait {
		return fmt.Errorf("Key %s is locked by id=%s", m.builder.key, m.id)
	}

	log.Errorf(err, "Key %s is locked, waiting for other node releases it, id=%s", m.builder.key, m.id)

	ctx, cancel := context.WithTimeout(m.builder.ctx, time.Duration(m.builder.ttl)*time.Second)
	gopool.Go(func(context.Context) {
		defer cancel()
		err := backend.Registry().Watch(ctx,
			registry.WithStrKey(m.builder.key),
			registry.WithWatchCallback(
				func(message string, evt *registry.PluginResponse) error {
					if evt != nil && evt.Action == registry.Delete {
						// break this for-loop, and try to create the node again.
						return fmt.Errorf("Lock released")
					}
					return nil
				}))
		if err != nil {
			log.Warnf("%s, key=%s, id=%s", err.Error(), m.builder.key, m.id)
		}
	})
	select {
	case <-ctx.Done():
		return ctx.Err() // 可以重新尝试获取锁
	case <-m.builder.ctx.Done():
		cancel()
		return m.builder.ctx.Err() // 机制错误，不应该超时的
	}
}

func (m *DLock) Unlock() (err error) {
	defer func() {
		if !IsDebug {
			m.builder.mutex.Unlock()
		}
	}()

	opts := []registry.PluginOpOption{
		registry.DEL,
		registry.WithStrKey(m.builder.key)}

	for i := 1; i <= DEFAULT_RETRY_TIMES; i++ {
		_, err = backend.Registry().Do(m.builder.ctx, opts...)
		if err == nil {
			log.Infof("Delete lock OK, key=%s, id=%s", m.builder.key, m.id)
			return nil
		}
		log.Errorf(err, "Delete lock failed, key=%s, id=%s", m.builder.key, m.id)
		e, ok := err.(client.Error)
		if ok && e.Code == client.ErrorCodeKeyNotFound {
			return nil
		}
	}
	return err
}

func Lock(key string, wait bool) (*DLock, error) {
	globalMux.Lock()
	lc, ok := globalMap[key]
	if !ok {
		lc = NewLockFactory(fmt.Sprintf("%s%s", ROOT_PATH, key), -1)
		globalMap[key] = lc
	}
	globalMux.Unlock()
	return lc.NewDLock(wait)
}
