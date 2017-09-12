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
package etcdsync

import (
	"fmt"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"io"
	"os"
	"sync"
	"time"
)

const (
	defaultTTL = 60
	defaultTry = 3
	ROOT_PATH  = "/cse/etcdsync"
)

// A Mutex is a mutual exclusion lock which is distributed across a cluster.
type LockerFactory struct {
	key    string
	ctx    context.Context
	ttl    int64
	mutex  *sync.Mutex
	logger io.Writer
}

type Locker struct {
	builder *LockerFactory
	id      string
}

var (
	globalMap map[string]*LockerFactory
	globalMux sync.Mutex
	IsDebug   bool
	hostname  string
	pid       int
)

func init() {
	globalMap = make(map[string]*LockerFactory)
	IsDebug = false

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}
	pid = os.Getpid()
}

// New creates a Mutex with the given key which must be the same
// across the cluster nodes.
// machines are the ectd cluster addresses
func New(key string, ttl int64) *LockerFactory {
	if len(key) == 0 {
		return nil
	}
	if ttl < 1 {
		ttl = defaultTTL
	}

	return &LockerFactory{
		key:   key,
		ctx:   context.Background(),
		ttl:   ttl,
		mutex: new(sync.Mutex),
	}
}

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *LockerFactory) Lock() (l *Locker, err error) {
	if !IsDebug {
		m.mutex.Lock()
	}
	l = &Locker{
		builder: m,
		id:      fmt.Sprintf("%v-%v-%v", hostname, pid, time.Now().Format("20060102-15:04:05.999999999")),
	}
	for try := 1; try <= defaultTry; try++ {
		err = l.Lock()
		if err == nil {
			return l, nil
		}

		if try <= defaultTry {
			util.Logger().Warnf(err, "Try to lock key %s again, id=%s", m.key, l.id)
		} else {
			util.Logger().Errorf(err, "Lock key %s failed, id=%s", m.key, l.id)
		}
	}
	if !IsDebug {
		m.mutex.Unlock()
	}
	return l, err
}

func (m *Locker) ID() string {
	return m.id
}

func (m *Locker) Lock() error {
	opts := []registry.PluginOpOption{
		registry.WithStrKey(m.builder.key),
		registry.WithStrValue(m.id)}

	for {
		util.Logger().Infof("Trying to create a lock: key=%s, id=%s", m.builder.key, m.id)

		putOpts := opts
		if m.builder.ttl > 0 {
			leaseID, err := registry.GetRegisterCenter().LeaseGrant(m.builder.ctx, m.builder.ttl)
			if err != nil {
				return err
			}
			putOpts = append(opts, registry.WithLease(leaseID))
		}
		success, err := registry.GetRegisterCenter().PutNoOverride(m.builder.ctx, putOpts...)
		if err == nil && success {
			util.Logger().Infof("Create Lock OK, key=%s, id=%s", m.builder.key, m.id)
			return nil
		}
		util.Logger().Warnf(err, "Key %s is locked, waiting for other node releases it, id=%s", m.builder.key, m.id)

		ctx, cancel := context.WithTimeout(m.builder.ctx, defaultTTL*time.Second)
		go func() {
			err := registry.GetRegisterCenter().Watch(ctx,
				registry.GET,
				registry.WithStrKey(m.builder.key),
				registry.WithWatchCallback(
					func(message string, evt *registry.PluginResponse) error {
						if evt != nil && evt.Action == registry.Delete {
							// break this for-loop, and try to create the node again.
							return fmt.Errorf("Lock released, id=%s", m.id)
						}
						return nil
					}))
			if err != nil {
				util.Logger().Errorf(nil, "%s, key=%s, id=%s", err.Error(), m.builder.key, m.id)
			}
			cancel()
		}()
		select {
		case <-ctx.Done():
			continue // 可以重新尝试获取锁
		case <-m.builder.ctx.Done():
			cancel()
			return m.builder.ctx.Err() // 机制错误，不应该超时的
		}
	}
}

// Unlock unlocks m.
// It is a run-time error if m is not locked on entry to Unlock.
//
// A locked Mutex is not associated with a particular goroutine.
// It is allowed for one goroutine to lock a Mutex and then
// arrange for another goroutine to unlock it.
func (m *Locker) Unlock() (err error) {
	opts := []registry.PluginOpOption{
		registry.DEL,
		registry.WithStrKey(m.builder.key)}

	for i := 1; i <= defaultTry; i++ {
		_, err = registry.GetRegisterCenter().Do(m.builder.ctx, opts...)
		if err == nil {
			if !IsDebug {
				m.builder.mutex.Unlock()
			}
			util.Logger().Infof("Delete lock OK, key=%s, id=%s", m.builder.key, m.id)
			return nil
		}
		util.Logger().Errorf(err, "Delete lock falied, key=%s, id=%s", m.builder.key, m.id)
		e, ok := err.(client.Error)
		if ok && e.Code == client.ErrorCodeKeyNotFound {
			if !IsDebug {
				m.builder.mutex.Unlock()
			}
			return nil
		}
	}
	if !IsDebug {
		m.builder.mutex.Unlock()
	}
	return err
}

func Lock(key string) (*Locker, error) {
	globalMux.Lock()
	lc, ok := globalMap[key]
	if !ok {
		lc = New(fmt.Sprintf("%s%s", ROOT_PATH, key), -1)
		globalMap[key] = lc
	}
	globalMux.Unlock()
	return lc.Lock()
}
