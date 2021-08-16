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
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core/backend"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/service/lock"
	"github.com/coreos/etcd/client"
)

const (
	DefaultLockTTL    = 60
	DefaultRetryTimes = 3
	RootPath          = "/cse/etcdsync"

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
	IsDebug  bool
	hostname = util.HostName()
	pid      = os.Getpid()
)

func init() {
	// TODO: set logger
	// TODO: register storage plugin to plugin manager
}

type DataSource struct{}

func NewDataSource() *DataSource {
	// TODO: construct a reasonable DataSource instance
	log.Warnf("auth data source enable etcd mode")

	inst := &DataSource{}
	// TODO: deal with exception
	if err := inst.initialize(); err != nil {
		return inst
	}
	return inst
}

func (ds *DataSource) initialize() error {
	// TODO: init DataSource members
	return nil
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
		err = backend.Registry().LeaseRevoke(m.ctx, leaseID)
		if err != nil {
			return err
		}
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

	for i := 1; i <= DefaultRetryTimes; i++ {
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

func (ds *DataSource) NewDLock(key string, ttl int64, wait bool) (lock.DLock, error) {
	if len(key) == 0 {
		return nil, nil
	}
	key = fmt.Sprintf("%s%s", RootPath, key)
	if ttl < 1 {
		ttl = DefaultLockTTL
	}

	now := time.Now()
	l := &DLock{
		key:      key,
		ctx:      context.Background(),
		ttl:      ttl,
		id:       fmt.Sprintf("%v-%v-%v", hostname, pid, now.Format("20060102-15:04:05.999999999")),
		createAt: now,
		mutex:    &sync.Mutex{},
	}
	var err error
	for try := 1; try <= DefaultRetryTimes; try++ {
		err = l.Lock(wait)
		if err == nil {
			return l, nil
		}
		if !wait {
			break
		}
	}
	// failed
	log.Errorf(err, "Lock key %s failed, id=%s", l.key, l.id)
	return nil, err
}
