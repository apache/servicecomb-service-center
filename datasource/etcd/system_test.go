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
package etcd_test

import (
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdminService_Dump(t *testing.T) {
	t.Log("execute 'dump' operation,when get all,should be passed")
	var cache model.Cache
	datasource.Instance().DumpCache(getContext(), &cache)
	assert.Equal(t, len(cache.Indexes), len(cache.Microservices))
}
