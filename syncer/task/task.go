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

package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/apache/servicecomb-service-center/pkg/log"
)

var taskMgr = map[string]generator{}

type generator func(params map[string]string) (Tasker, error)

// Tasker interface
type Tasker interface {
	Run(ctx context.Context)
	Handle(handler func())
}

// RegisterTasker register an tasker to manager
func RegisterTasker(name string, fn generator) {
	if _, ok := taskMgr[name]; ok {
		log.Warn(fmt.Sprintf("task generator is already exist, name = %s", name))
	}
	taskMgr[name] = fn
}

// GenerateTasker generate an tasker by name from manager
func GenerateTasker(name string, ops ...Option) (Tasker, error) {
	fn, ok := taskMgr[name]
	if !ok {
		err := errors.New("trigger generator is not found")
		log.Error(fmt.Sprintf("name = %s", name), err)
		return nil, err
	}
	return fn(toMap(ops...))
}

// Option task option
type Option func(map[string]string)

// WithAddKV wrap the key and value to an option
func WithAddKV(key, value string) Option {
	return func(m map[string]string) { m[key] = value }
}

func toMap(ops ...Option) map[string]string {
	m := make(map[string]string, len(ops))
	for _, op := range ops {
		op(m)
	}
	return m
}
