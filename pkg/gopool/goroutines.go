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
package gopool

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/log"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"golang.org/x/net/context"
	"sync"
	"time"
)

var GlobalConfig = Configure()

var defaultGo *Pool

func init() {
	defaultGo = New(context.Background())
}

type Config struct {
	Concurrent  int
	IdleTimeout time.Duration
}

func (c *Config) Workers(max int) *Config {
	c.Concurrent = max
	return c
}

func (c *Config) Idle(time time.Duration) *Config {
	c.IdleTimeout = time
	return c
}

func Configure() *Config {
	return &Config{
		Concurrent:  1000,
		IdleTimeout: 60 * time.Second,
	}
}

type Pool struct {
	Cfg *Config

	// job context
	ctx    context.Context
	cancel context.CancelFunc
	// pending is the chan to block Pool.Do() when go pool is full
	pending chan func(ctx context.Context)
	// workers is the counter of the worker
	workers chan struct{}

	mux    sync.RWMutex
	wg     sync.WaitGroup
	closed bool
}

func (g *Pool) execute(f func(ctx context.Context)) {
	defer log.Recover()
	f(g.ctx)
}

func (g *Pool) Do(f func(context.Context)) *Pool {
	defer log.Recover()
	select {
	case g.pending <- f: // block if workers are busy
	case g.workers <- struct{}{}:
		g.wg.Add(1)
		go g.loop(f)
	}
	return g
}

func (g *Pool) loop(f func(context.Context)) {
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
			util.ResetTimer(timer, g.Cfg.IdleTimeout)
		}
	}
}

// Close will call context.Cancel(), so all goroutines maybe exit when job does not complete
func (g *Pool) Close(grace bool) {
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
func (g *Pool) Done() {
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

func New(ctx context.Context, cfgs ...*Config) *Pool {
	ctx, cancel := context.WithCancel(ctx)
	if len(cfgs) == 0 {
		cfgs = append(cfgs, GlobalConfig)
	}
	cfg := cfgs[0]
	gr := &Pool{
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

func CloseAndWait() {
	defaultGo.Close(true)
	log.Debugf("all goroutines exited")
}
