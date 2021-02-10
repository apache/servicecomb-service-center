/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package sd

import (
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	options := Options{
		Key: "",
	}
	assert.Empty(t, options, "config is empty")

	options1 := options.SetTable("configKey")
	assert.Equal(t, "configKey", options1.Key,
		"contain key after method WithTable")

	assert.Equal(t, 0, options1.InitSize,
		"init size is zero")

	mongoEventFunc = mongoEventFuncGet()

	out := options1.String()
	assert.NotNil(t, out,
		"method String return not after methods")
}

var mongoEventFunc MongoEventFunc

func mongoEventFuncGet() MongoEventFunc {
	fun := func(evt MongoEvent) {
		evt.DocumentID = "DocumentID has changed"
		evt.ResourceID = "BusinessID has changed"
		evt.Value = 2
		evt.Type = discovery.EVT_UPDATE
		log.Info("in event func")
	}
	return fun
}
