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

package util_test

import (
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWildcardMatch(t *testing.T) {
	t.Run("not regexp should match exactly", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("TestA", "TestA"))
		assert.False(t, util.WildcardMatch("TestA", "TestAB"))
		assert.False(t, util.WildcardMatch("TestA", "BTestA"))
	})

	t.Run("start with * should match true", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("*estA", "TestA"))
	})

	t.Run("contain * should match true", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("Tes*A", "TestA"))
	})

	t.Run("end with * should match true", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("Test*", "TestA"))
	})

	t.Run("end with * should match true", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("Test*", "Test"))
	})

	t.Run("not match should return false", func(t *testing.T) {
		assert.False(t, util.WildcardMatch("Test*", "DevA"))
	})

	t.Run("contain other should return true", func(t *testing.T) {
		assert.True(t, util.WildcardMatch("^(Test)?[a-z]*\\w+$", "^(Test)?[a-z]\\w+$"))
		assert.True(t, util.WildcardMatch("^(Test)?[a-z]*\\w+$", "^(Test)?[a-z]A\\w+$"))
	})
}
