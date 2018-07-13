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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
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

func syncFunc(i int) {
	if i >= count {
		return
	}
	syncFunc(i + 1)
}

func BenchmarkInvocationOption(b *testing.B) {
	var (
		op   chain.InvocationOp
		f    = func(r chain.Result) {}
		opts = []chain.InvocationOption{chain.WithFunc(f), chain.WithAsyncFunc(f)}
	)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, opt := range opts {
				op = opt(op)
			}
		}
	})
	b.ReportAllocs()
	// 50000000	        26.8 ns/op	       0 B/op	       0 allocs/op
}

func BenchmarkChain(b *testing.B) {
	var (
		ctx = util.NewStringContext(context.Background())
		f   = func(r chain.Result) {}
	)

	b.N = times
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			inv := chain.NewInvocation(ctx, chain.NewChain("_bench_chain_", chain.Handlers("_bench_handlers_")))
			inv.Invoke(f)
		}
	})
	b.ReportAllocs()
	// 1000000	      6607 ns/op	      80 B/op	       1 allocs/op
}

func BenchmarkSync(b *testing.B) {
	b.N = times
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			syncFunc(0)
		}
	})
	b.ReportAllocs()
	// 1000000	        46.9 ns/op	       0 B/op	       0 allocs/op
}
