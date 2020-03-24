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
	"testing"

	"github.com/stretchr/testify/assert"
)

// empty struct
type empty struct{}

// newEmpty returns an empty task
func newEmpty(params map[string]string) (Tasker, error) {
	return &empty{}, nil
}

// Run task
func (t *empty) Run(ctx context.Context) {}

// Handle task trigger
func (t *empty) Handle(handler func()) {}

func TestTasker(t *testing.T) {
	taskName := "empty"
	_, err := GenerateTasker(taskName)
	assert.NotNil(t, err)

	RegisterTasker(taskName, newEmpty)
	RegisterTasker(taskName, newEmpty)

	_, err = GenerateTasker(taskName, WithAddKV("test", "test"))
	assert.Nil(t, err)
}
