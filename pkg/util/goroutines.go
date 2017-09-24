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
package util

import "sync"

type GoRoutine struct {
	stopCh chan struct{}
	wg     *sync.WaitGroup
	mux    sync.RWMutex
	once   sync.Once
	closed bool
}

func (g *GoRoutine) Init(stopCh chan struct{}) {
	g.once.Do(func() {
		g.wg = &sync.WaitGroup{}
		g.stopCh = stopCh
	})
}

func (g *GoRoutine) StopCh() <-chan struct{} {
	return g.stopCh
}

func (g *GoRoutine) Do(f func(<-chan struct{})) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		f(g.StopCh())
	}()
}

func (g *GoRoutine) Close(wait bool) {
	g.mux.Lock()
	defer g.mux.Unlock()
	if g.closed {
		return
	}
	g.closed = true
	close(g.stopCh)
	if wait {
		g.Wait()
	}
}

func (g *GoRoutine) Wait() {
	g.wg.Wait()
}

var defaultGo GoRoutine

func init() {
	GoInit()
}

func Go(f func(<-chan struct{})) {
	defaultGo.Do(f)
}

func GoInit() {
	defaultGo.Init(make(chan struct{}))
}

func GoCloseAndWait() {
	defaultGo.Close(true)
}

func NewGo(stopCh chan struct{}) *GoRoutine {
	gr := &GoRoutine{}
	gr.Init(stopCh)
	return gr
}
