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
	"time"
)

const (
	CtxDomain        = "domain"
	CtxProject       = "project"
	CtxTargetDomain  = "target-domain"
	CtxTargetProject = "target-project"
)

type StringContext struct {
	parentCtx context.Context
	kv        *ConcurrentMap
}

func (c *StringContext) Deadline() (deadline time.Time, ok bool) {
	return c.parentCtx.Deadline()
}

func (c *StringContext) Done() <-chan struct{} {
	return c.parentCtx.Done()
}

func (c *StringContext) Err() error {
	return c.parentCtx.Err()
}

func (c *StringContext) Value(key interface{}) interface{} {
	k, ok := key.(string)
	if !ok {
		return c.parentCtx.Value(key)
	}
	v, ok := c.kv.Get(k)
	if !ok {
		return FromContext(c.parentCtx, k)
	}
	return v
}

func (c *StringContext) SetKV(key string, val interface{}) {
	c.kv.Put(key, val)
}

func NewStringContext(ctx context.Context) *StringContext {
	strCtx, ok := ctx.(*StringContext)
	if !ok {
		strCtx = &StringContext{
			parentCtx: ctx,
			kv:        NewConcurrentMap(0),
		}
	}
	return strCtx
}

func SetContext(ctx context.Context, key string, val interface{}) context.Context {
	strCtx := NewStringContext(ctx)
	strCtx.SetKV(key, val)
	return strCtx
}

func CloneContext(ctx context.Context) context.Context {
	old, ok := ctx.(*StringContext)
	if !ok {
		return &StringContext{
			parentCtx: ctx,
			kv:        NewConcurrentMap(0),
		}
	}

	strCtx := &StringContext{
		parentCtx: ctx,
		kv:        NewConcurrentMap(0),
	}

	old.kv.ForEach(func(item MapItem) bool {
		strCtx.kv.Put(item.Key, item.Value)
		return true
	})
	return strCtx
}

func FromContext(ctx context.Context, key string) interface{} {
	if v := ctx.Value(key); v != nil {
		return v
	}
	return FromMetadata(ctx, key)
}

func SetRequestContext(r *http.Request, key string, val interface{}) *http.Request {
	ctx := r.Context()
	ctx = SetContext(ctx, key, val)
	if ctx != r.Context() {
		nr := r.WithContext(ctx)
		*r = *nr
	}
	return r
}

func ParseDomainProject(ctx context.Context) string {
	return ParseDomain(ctx) + "/" + ParseProject(ctx)
}

func ParseTargetDomainProject(ctx context.Context) string {
	return ParseTargetDomain(ctx) + "/" + ParseTargetProject(ctx)
}

func ParseDomain(ctx context.Context) string {
	v, ok := FromContext(ctx, CtxDomain).(string)
	if !ok {
		return ""
	}
	return v
}

func ParseTargetDomain(ctx context.Context) string {
	v, _ := FromContext(ctx, CtxTargetDomain).(string)
	if len(v) == 0 {
		return ParseDomain(ctx)
	}
	return v
}

func ParseProject(ctx context.Context) string {
	v, ok := FromContext(ctx, CtxProject).(string)
	if !ok {
		return ""
	}
	return v
}

func ParseTargetProject(ctx context.Context) string {
	v, _ := FromContext(ctx, CtxTargetProject).(string)
	if len(v) == 0 {
		return ParseProject(ctx)
	}
	return v
}

func SetDomain(ctx context.Context, domain string) context.Context {
	return SetContext(ctx, CtxDomain, domain)
}

func SetProject(ctx context.Context, project string) context.Context {
	return SetContext(ctx, CtxProject, project)
}

func SetTargetDomain(ctx context.Context, domain string) context.Context {
	return SetContext(ctx, CtxTargetDomain, domain)
}

func SetTargetProject(ctx context.Context, project string) context.Context {
	return SetContext(ctx, CtxTargetProject, project)
}

func SetDomainProject(ctx context.Context, domain string, project string) context.Context {
	return SetProject(SetDomain(ctx, domain), project)
}

func SetTargetDomainProject(ctx context.Context, domain string, project string) context.Context {
	return SetTargetProject(SetTargetDomain(ctx, domain), project)
}
