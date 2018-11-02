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
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestFromMetadata(t *testing.T) {
	v := FromMetadata(context.Background(), "a")
	if v != "" {
		t.Fatalf("TestFromMetadata failed")
	}
	ctx := context.WithValue(context.Background(), "a", "b")
	v = FromMetadata(ctx, "a")
	if v != "" {
		t.Fatalf("TestFromMetadata failed")
	}
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})
	v = FromMetadata(ctx, "a")
	if v != "" {
		t.Fatalf("TestFromMetadata failed")
	}
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"a": []string{}})
	v = FromMetadata(ctx, "a")
	if v != "" {
		t.Fatalf("TestFromMetadata failed")
	}
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"a": []string{"b", "c"}})
	v = FromMetadata(ctx, "a")
	if v != "b" {
		t.Fatalf("TestFromMetadata failed")
	}

	// clone
	cloneCtx := CloneContext(ctx)
	v = FromMetadata(cloneCtx, "a")
	if v != "b" {
		t.Fatalf("TestFromMetadata failed")
	}
}
