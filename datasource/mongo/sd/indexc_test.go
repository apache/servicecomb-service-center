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

package sd_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/stretchr/testify/assert"
)

func TestIndexCache(t *testing.T) {
	indexCache := sd.NewIndexCache()
	assert.NotNil(t, indexCache)
	indexCache.Put("index1", "doc1")
	assert.Equal(t, []string{"doc1"}, indexCache.Get("index1"))
	indexCache.Put("index1", "doc2")
	assert.Len(t, indexCache.Get("index1"), 2)
	indexCache.Delete("index1", "doc1")
	assert.Len(t, indexCache.Get("index1"), 1)
	indexCache.Delete("index1", "doc2")
	assert.Nil(t, indexCache.Get("index1"))
}
