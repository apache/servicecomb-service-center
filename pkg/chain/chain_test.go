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
package chain

import (
	"context"
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"testing"
	"time"
)

const (
	times = 1000000
	count = 10
)

func init() {
	for i := 0; i < count; i++ {
		RegisterHandler("_bench_handlers_", &handler{})
	}
}

type handler struct {
}

func (h *handler) Handle(i *Invocation) {
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
		op   InvocationOp
		f    = func(r Result) {}
		opts = []InvocationOption{WithFunc(f), WithAsyncFunc(f)}
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
		f   = func(r Result) {}
	)

	b.N = times
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inv := NewInvocation(ctx, NewChain("_bench_chain_", Handlers("_bench_handlers_")))
		inv.Invoke(f)

	}
	b.ReportAllocs()
	// 1000000	       735 ns/op	      80 B/op	       1 allocs/op
}

type mockHandler struct {
}

func (h *mockHandler) Handle(i *Invocation) {
	x := i.Context().Value("x").(int)
	switch x {
	case 1:
		i.Success(x)
	case 2:
		i.Fail(errors.New("error"), x)
	case 3:
		i.Next(WithFunc(func(r Result) {
			i.WithContext("x", x*x)
		}))
	case 4:
		i.Next(WithAsyncFunc(func(r Result) {
			i.WithContext("x", x*x)
			ch, _ := i.Context().Value("ch").(chan struct{})
			ch <- struct{}{}
		}))
	case 5:
		panic(errors.New("error"))
	case 6:
		i.Next(WithFunc(func(r Result) {
			panic(errors.New("error"))
		}))
	case 7:
		i.Next(WithAsyncFunc(func(r Result) {
			panic(errors.New("error"))
		}))
	default:
		i.WithContext("x", x-1)
		i.Next()
	}
}

func TestChain_Next(t *testing.T) {
	h := &mockHandler{}
	hs := Handlers("test")
	hs = []Handler{h, h, h}

	x := 0
	ch := NewChain("test", hs)
	if ch.Name() != "test" {
		t.Fatalf("TestChain_Next")
	}
	i := NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if !r.OK || i.Context().Value("x").(int) != -len(hs) {
			t.Fatalf("TestChain_Next")
		}
	})

	x = 1
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if !r.OK || r.Args[0].(int) != x {
			t.Fatalf("TestChain_Next")
		}
	})

	x = 2
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if r.OK || r.Args[0].(int) != x {
			t.Fatalf("TestChain_Next")
		}
	})

	x = 3
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if !r.OK || i.Context().Value("x").(int) != x {
			t.Fatalf("TestChain_Next")
		}
	})
	if i.Context().Value("x").(int) != x*x {
		t.Fatalf("TestChain_Next")
	}

	x = 4 // async call back
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.WithContext("ch", make(chan struct{}))
	i.Invoke(func(r Result) {
		if !r.OK || i.Context().Value("x").(int) != x {
			t.Fatalf("TestChain_Next")
		}
	})
	<-i.Context().Value("ch").(chan struct{})
	if i.Context().Value("x").(int) != x*x {
		t.Fatalf("TestChain_Next")
	}

	x = 5
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if r.OK || r.Err == nil {
			t.Fatalf("TestChain_Next")
		}
	})

	x = 6
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if !r.OK || r.Err != nil {
			t.Fatalf("TestChain_Next")
		}
	})

	x = 7 // async call back
	ch = NewChain("test", hs)
	i = NewInvocation(context.Background(), ch)
	i.WithContext("x", x)
	i.Invoke(func(r Result) {
		if !r.OK || r.Err != nil {
			t.Fatalf("TestChain_Next")
		}
	})
	<-time.After(500 * time.Millisecond)
}
