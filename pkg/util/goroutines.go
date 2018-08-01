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
package util

import (
	"golang.org/x/net/context"
	"sync"
	"time"
)

var GlobalPoolConfig = PoolConfigure()

var defaultGo *GoRoutine

func init() {
	defaultGo = NewGo(context.Background())
}

type PoolConfig struct {
	Concurrent  int
	IdleTimeout time.Duration
}

func (c *PoolConfig) Workers(max int) *PoolConfig {
	c.Concurrent = max
	return c
}

func (c *PoolConfig) Idle(time time.Duration) *PoolConfig {
	c.IdleTimeout = time
	return c
}

func PoolConfigure() *PoolConfig {
	return &PoolConfig{
		Concurrent:  1000,
		IdleTimeout: 60 * time.Second,
	}
}

type GoRoutine struct {
	Cfg *PoolConfig

	// job context
	ctx    context.Context
	cancel context.CancelFunc
	// pending is the chan to block GoRoutine.Do() when go pool is full
	pending chan func(ctx context.Context)
	// workers is the counter of the worker
	workers chan struct{}

	mux    sync.RWMutex
	wg     sync.WaitGroup
	closed bool
}

func (g *GoRoutine) execute(f func(ctx context.Context)) {
	defer RecoverAndReport()
	f(g.ctx)
}

func (g *GoRoutine) Do(f func(context.Context)) *GoRoutine {
	defer RecoverAndReport()
	select {
	case g.pending <- f:
	case g.workers <- struct{}{}:
		g.wg.Add(1)
		go g.loop(f)
	}
	return g
}

func (g *GoRoutine) loop(f func(context.Context)) {
	defer g.wg.Done()
	defer func() { <-g.workers }()

	timer := time.NewTimer(g.Cfg.IdleTimeout)
	defer timer.Stop()
	for {
		g.execute(f)

		select {
		case <-timer.C:
			return
		case f = <-g.pending:
			if f == nil {
				return
			}
			ResetTimer(timer, g.Cfg.IdleTimeout)
		}
	}
}

// Close will call context.Cancel(), so all goroutines maybe exit when job does not complete
func (g *GoRoutine) Close(grace bool) {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)
	g.cancel()
	if grace {
		g.wg.Wait()
	}
}

// Done will wait for all goroutines complete the jobs and then close the pool
func (g *GoRoutine) Done() {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pending)
	close(g.workers)
	g.wg.Wait()
}

func NewGo(ctx context.Context, cfgs ...*PoolConfig) *GoRoutine {
	ctx, cancel := context.WithCancel(ctx)
	if len(cfgs) == 0 {
		cfgs = append(cfgs, GlobalPoolConfig)
	}
	cfg := cfgs[0]
	gr := &GoRoutine{
		Cfg:     cfg,
		ctx:     ctx,
		cancel:  cancel,
		pending: make(chan func(context.Context)),
		workers: make(chan struct{}, cfg.Concurrent),
	}
	return gr
}

func Go(f func(context.Context)) {
	defaultGo.Do(f)
}

func GoCloseAndWait() {
	defaultGo.Close(true)
	Logger().Debugf("all goroutines exited")
}
