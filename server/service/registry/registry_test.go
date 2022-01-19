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

package registry_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/service/registry"
	"github.com/stretchr/testify/assert"
)

func TestSelfRegister(t *testing.T) {
	t.Run("self register after upgrade sc version, should be ok", func(t *testing.T) {
		oldServiceID := core.Service.ServiceId
		oldVersion := core.Service.Version
		defer func() {
			core.Service.ServiceId = oldServiceID
			core.Service.Version = oldVersion
		}()

		core.Service.ServiceId = ""
		core.Service.Version = "0.0." + strconv.Itoa(rand.Intn(100))
		err := registry.SelfRegister(context.Background())
		assert.NoError(t, err)
	})
}
