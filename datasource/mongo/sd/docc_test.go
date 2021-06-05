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

	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
)

var docCache *sd.DocStore

func init() {
	docCache = sd.NewDocStore()
}

func TestDocCache(t *testing.T) {
	t.Run("init docCache,should pass", func(t *testing.T) {
		docCache = sd.NewDocStore()
		assert.NotNil(t, docCache)
		assert.Nil(t, docCache.Get("id1"))
	})
	t.Run("update&&delete docCache, should pass", func(t *testing.T) {
		docCache.Put("id1", "doc1")
		assert.Equal(t, "doc1", docCache.Get("id1").(string))
		assert.Equal(t, 1, docCache.Size())
		docCache.Put("id2", "doc2")
		assert.Equal(t, 2, docCache.Size())
		docCache.DeleteDoc("id2")
		assert.Equal(t, 1, docCache.Size())
	})
}
