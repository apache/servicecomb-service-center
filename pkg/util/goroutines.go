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

var GlobalPoolConfig = &PoolConfig{
	PutTimeout:  50 * time.Millisecond,
	IdleTimeout: 30 * time.Second,
}

type PoolConfig struct {
	Max         int
	PutTimeout  time.Duration
	IdleTimeout time.Duration
}

type GoRoutine struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mux    sync.RWMutex
	closed bool
	pool   chan func(ctx context.Context)
}

func (g *GoRoutine) execute(f func(ctx context.Context)) {
	defer RecoverAndReport()
	f(g.ctx)
}

func (g *GoRoutine) Do(f func(context.Context)) {
	defer RecoverAndReport()
	for {
		select {
		case g.pool <- f:
			return
		case <-time.After(GlobalPoolConfig.PutTimeout):
			go g.loop()
		}
	}
}

func (g *GoRoutine) loop() {
	g.wg.Add(1)
	defer g.wg.Done()
	for {
		select {
		case <-time.After(GlobalPoolConfig.IdleTimeout):
			return
		case f, ok := <-g.pool:
			if !ok {
				return
			}
			g.execute(f)
		}
	}
}

func (g *GoRoutine) Close(wait bool) {
	g.mux.Lock()
	if g.closed {
		g.mux.Unlock()
		return
	}
	g.closed = true
	g.mux.Unlock()

	close(g.pool)
	g.cancel()
	if wait {
		g.Wait()
	}
}

func (g *GoRoutine) Wait() {
	g.wg.Wait()
}

var defaultGo *GoRoutine

func init() {
	defaultGo = NewGo(context.Background())
}

func Go(f func(context.Context)) {
	defaultGo.Do(f)
}

func GoCloseAndWait() {
	defaultGo.Close(true)
	Logger().Debugf("all goroutines exited")
}

func NewGo(ctx context.Context) *GoRoutine {
	ctx, cancel := context.WithCancel(ctx)
	gr := &GoRoutine{
		ctx:    ctx,
		cancel: cancel,
		pool:   make(chan func(context.Context)),
	}
	return gr
}
