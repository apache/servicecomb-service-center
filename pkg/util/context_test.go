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
	"context"
	"net/http"
	"testing"
)

func TestSetContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "z", "z")
	ctx = context.WithValue(ctx, 1, 1)
	ctx = SetDomainProject(ctx, "x", "y")
	if ParseDomainProject(ctx) != "x/y" {
		t.Fatalf("TestSetContext failed")
	}
	if ctx.Value("z") != "z" {
		t.Fatalf("TestSetContext failed")
	}
	if ctx.Value(1) != 1 {
		t.Fatalf("TestSetContext failed")
	}

	nctx := CloneContext(ctx)
	if nctx == ctx {
		t.Fatalf("TestSetContext failed")
	}
	if ParseDomainProject(nctx) != "x/y" {
		t.Fatalf("TestSetContext failed")
	}
	nctx = SetDomainProject(nctx, "x", "x")
	if ParseDomainProject(ctx) != "x/y" {
		t.Fatalf("TestSetContext failed")
	}
	if ParseDomainProject(nctx) != "x/x" {
		t.Fatalf("TestSetContext failed")
	}
	if nctx.Value("z") != "z" {
		t.Fatalf("TestSetContext failed")
	}
	if nctx.Value(1) != 1 {
		t.Fatalf("TestSetContext failed")
	}

	ctx = SetTargetDomainProject(ctx, "x", "x")
	if ParseTargetDomainProject(ctx) != "x/x" {
		t.Fatalf("TestSetContext failed")
	}
	if ctx.Value("z") != "z" {
		t.Fatalf("TestSetContext failed")
	}
	if ctx.Value(1) != 1 {
		t.Fatalf("TestSetContext failed")
	}

	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:30100", nil)
	ctx = req.Context()
	SetRequestContext(req, "x", "y")
	if req.Context() == ctx {
		t.Fatalf("TestSetContext failed")
	}
}
