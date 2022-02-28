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

package sd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeStore_Initialize(t *testing.T) {
	s := TypeStore{}
	s.Initialize()
	assert.NotNil(t, s.ready)
	assert.NotNil(t, s.goroutine)
	assert.Equal(t, false, s.isClose)
}

func TestTypeStore_Ready(t *testing.T) {
	s := TypeStore{}
	s.Initialize()
	c := s.Ready()
	assert.NotNil(t, c)
}

func TestTypeStore_Stop(t *testing.T) {
	t.Run("when closed", func(t *testing.T) {
		s := TypeStore{
			isClose: true,
		}
		s.Initialize()
		s.Stop()
		assert.Equal(t, true, s.isClose)
	})

	t.Run("when not closed", func(t *testing.T) {
		s := TypeStore{
			isClose: false,
		}
		s.Initialize()
		s.Stop()
		assert.Equal(t, true, s.isClose)
	})
}
