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
	"github.com/apache/incubator-servicecomb-service-center/pkg/chain"
	"testing"
)

const (
	times = 1000000
	count = 100
)

func init() {
	for i := 0; i < count; i++ {
		chain.RegisterHandler("_bench_handlers_", &handler{})
	}
}

type handler struct {
}

func (h *handler) Handle(i *chain.Invocation) {
	i.Next()
}

func syncFunc() {
	for i := 0; i < count; i++ {
	}
}

func BenchmarkChain(b *testing.B) {
	var (
		ctx = context.Background()
		f   = func(r chain.Result) {}
	)
	b.N = times
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			inv := chain.NewInvocation(ctx, chain.NewChain("_bench_chain_", chain.Handlers("_bench_handlers_")))
			inv.Next(chain.WithFunc(f))
		}
	})
	b.ReportAllocs()
}

func BenchmarkSync(b *testing.B) {
	b.N = times
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			syncFunc()
		}
	})
	b.ReportAllocs()
}
