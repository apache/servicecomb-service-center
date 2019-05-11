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
package events

import "context"

type contextEvent struct {
	kind string
	ctx  context.Context
}

// NewContextEvent new ContextEvent with kind and context
func NewContextEvent(kind string, ctx context.Context) ContextEvent {
	return &contextEvent{
		kind: kind,
		ctx:  ctx,
	}
}

// Type get event kind
func (e *contextEvent) Type() string {
	return e.kind
}

// Type get event context
func (e *contextEvent) Context() context.Context {
	return e.ctx
}
