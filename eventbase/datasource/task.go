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

package datasource

import (
	"context"

	"github.com/go-chassis/cari/sync"
)

// TaskDao provide api of Task entity
type TaskDao interface {
	// Create func is used for ut
	Create(ctx context.Context, task *sync.Task) (*sync.Task, error)
	Update(ctx context.Context, task *sync.Task) error
	Delete(ctx context.Context, tasks ...*sync.Task) error
	List(ctx context.Context, domain string, project string, options ...TaskFindOption) ([]*sync.Task, error)
}
