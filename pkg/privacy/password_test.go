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

package privacy_test

import (
	"github.com/apache/servicecomb-service-center/pkg/privacy"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashPassword(t *testing.T) {
	h, _ := privacy.HashPassword("test")
	t.Log(h)
	mac, _ := privacy.ScryptPassword("test")
	t.Log(mac)

	t.Run("given old hash result, should be compatible", func(t *testing.T) {
		same := privacy.SamePassword(h, "test")
		assert.True(t, same)
	})

	sameMac := privacy.SamePassword(mac, "test")
	assert.True(t, sameMac)
}
