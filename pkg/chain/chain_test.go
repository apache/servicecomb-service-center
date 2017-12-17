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
package chain_test

import (
	"context"
	"github.com/ServiceComb/service-center/pkg/chain"
	"testing"
)

type handler struct {
}

func (h *handler) Handle(i *chain.Invocation) {
	counter := i.Context().Value("counter").(int)
	counter++
	i.WithContext("counter", counter)
	i.Next()
}

func syncFunc(ctx map[string]interface{}) {
	counter := ctx["counter"].(int)
	counter++
	ctx["counter"] = counter
}

func BenchmarkChain(b *testing.B) {
	b.N = 100
	if len(chain.Handlers("_bench_handlers_")) == 0 {
		for i := 0; i < b.N; i++ {
			chain.RegisterHandler("_bench_handlers_", &handler{})
		}
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch := make(chan struct{})
			inv := chain.NewInvocation(context.Background(),
				chain.NewChain("_bench_chain_", chain.Handlers("_bench_handlers_")))
			inv.WithContext("counter", 0)
			inv.Invoke(func(r chain.Result) { close(ch) })
			<-ch
		}
	})
	b.ReportAllocs()
}

func BenchmarkSync(b *testing.B) {
	b.N = 100
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := make(map[string]interface{}, 1)
			ctx["counter"] = 0
			for i := 0; i < b.N; i++ {
				syncFunc(ctx)
			}
		}
	})
	b.ReportAllocs()
}
