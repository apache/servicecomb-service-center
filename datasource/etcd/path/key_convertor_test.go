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
package path_test

import (
	"testing"

	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
)

func TestGetInfoFromKV(t *testing.T) {
	s, d := path.GetInfoFromSvcKV([]byte(path.GenerateServiceKey("a/b", "c")))
	assert.False(t, d != "a/b" || s != "c")

	s, d = path.GetInfoFromSvcKV([]byte("sdf"))
	assert.False(t, d != "" || s != "")

	var i string
	s, i, d = path.GetInfoFromInstKV([]byte(path.GenerateInstanceKey("a/b", "c", "d")))
	assert.False(t, d != "a/b" || s != "c" || i != "d")

	s, i, d = path.GetInfoFromInstKV([]byte("sdf"))
	assert.False(t, d != "" || s != "" || i != "")

	d = path.GetInfoFromDomainKV([]byte(path.GenerateDomainKey("a")))
	assert.False(t, d != "a")

	d = path.GetInfoFromDomainKV([]byte("sdf"))
	assert.False(t, d != "")

	p := ""
	d, p = path.GetInfoFromProjectKV([]byte(path.GenerateProjectKey("a", "b")))
	assert.False(t, d != "a" || p != "b")

	d, p = path.GetInfoFromProjectKV([]byte("sdf"))
	assert.False(t, d != "" || p != "")

	var r string
	s, r, d = path.GetInfoFromRuleKV([]byte(path.GenerateServiceRuleKey("a/b", "c", "d")))
	assert.False(t, d != "a/b" || s != "c" || r != "d")

	s, r, d = path.GetInfoFromRuleKV([]byte("sdf"))
	assert.False(t, d != "" || s != "" || r != "")

	s, d = path.GetInfoFromTagKV([]byte(path.GenerateServiceTagKey("a/b", "c")))
	assert.False(t, d != "a/b" || s != "c")

	s, d = path.GetInfoFromTagKV([]byte("sdf"))
	assert.False(t, d != "" || s != "")

	key := path.GetInfoFromSvcIndexKV([]byte(path.GenerateServiceIndexKey(&discovery.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
		Environment: "f",
		Alias:       "g",
	})))
	assert.False(t, key.Tenant != "a/b" ||
		key.AppId != "c" ||
		key.ServiceName != "d" ||
		key.Version != "e" ||
		key.Environment != "f" ||
		key.Alias != "")

	key = path.GetInfoFromSvcIndexKV([]byte("sdf"))
	assert.False(t, key != nil)

	d, s, m := path.GetInfoFromSchemaSummaryKV([]byte(path.GenerateServiceSchemaSummaryKey("a/b", "c", "d")))
	assert.False(t, m != "d" || s != "c" || d != "a/b")

	d, s, m = path.GetInfoFromSchemaSummaryKV([]byte("sdf"))
	assert.False(t, m != "" || s != "" || d != "")

	d, s, m = path.GetInfoFromSchemaKV([]byte(path.GenerateServiceSchemaKey("a/b", "c", "d")))
	assert.False(t, m != "d" || s != "c" || d != "a/b")

	d, s, m = path.GetInfoFromSchemaKV([]byte("sdf"))
	assert.False(t, m != "" || s != "" || d != "")

	u := ""
	s, d, u = path.GetInfoFromDependencyQueueKV([]byte(path.GenerateConsumerDependencyQueueKey("a/b", "c", "d")))
	assert.False(t, s != "c" || d != "a/b" || u != "d")

	s, d, u = path.GetInfoFromDependencyQueueKV([]byte("sdf"))
	assert.False(t, s != "" || d != "" || u != "")

	dt, k := path.GetInfoFromDependencyRuleKV([]byte(path.GenerateProviderDependencyRuleKey("a/b", &discovery.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "*",
	})))
	assert.False(t, dt != path.DepsProvider || k == nil || k.AppId != "" || k.ServiceName != "*")

	dt, k = path.GetInfoFromDependencyRuleKV([]byte(path.GenerateProviderDependencyRuleKey("a/b", &discovery.MicroServiceKey{
		Tenant:      "a/b",
		AppId:       "c",
		ServiceName: "d",
		Version:     "e",
	})))
	assert.False(t, dt != path.DepsProvider || k == nil || k.AppId != "c" || k.ServiceName != "d")

	dt, k = path.GetInfoFromDependencyRuleKV([]byte("abc"))
	assert.False(t, dt != "" || k != nil)
}
