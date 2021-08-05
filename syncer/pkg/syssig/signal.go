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

package syssig

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

var (
	once       sync.Once
	lock       sync.RWMutex
	handlerMap = map[os.Signal][]func(){
		syscall.SIGHUP:  {},
		syscall.SIGINT:  {},
		syscall.SIGKILL: {},
		syscall.SIGTERM: {},
	}
)

// Run start system signal listening
func Run(ctx context.Context) {
	once.Do(func() {
		listenSignals := make([]os.Signal, 0, 10)
		for key := range handlerMap {
			listenSignals = append(listenSignals, key)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, listenSignals...)

		select {
		case <-ctx.Done():

		case sig := <-sigChan:
			log.Info(fmt.Sprintf("system signal: %s", sig.String()))
			calls := callbacks(sig)
			for _, call := range calls {
				call()
			}
		}

	})
}

// AddSignalsHandler add system signal listener with types
func AddSignalsHandler(handler func(), signals ...os.Signal) error {
	for _, sig := range signals {
		lock.RLock()
		handlers, ok := handlerMap[sig]
		lock.RUnlock()
		if !ok {
			return fmt.Errorf("system signal %s is not notify", sig.String())
		}
		handlers = append(handlers, handler)
		lock.Lock()
		handlerMap[sig] = handlers
		lock.Unlock()
	}
	return nil
}

// callbacks Callback system listeners
func callbacks(signal os.Signal) []func() {
	lock.RLock()
	calls := handlerMap[signal]
	lock.RUnlock()
	return calls
}
