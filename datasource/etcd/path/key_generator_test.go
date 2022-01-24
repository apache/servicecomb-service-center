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

func TestGenerateRBACAccountKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/accounts/admin", path.GenerateRBACAccountKey("admin"))
}

func TestGenerateETCDProjectKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/projects/domain/project", path.GenerateProjectKey("domain", "project"))
}

func TestGenerateETCDDomainKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/domains/domain", path.GenerateDomainKey("domain"))
}

func TestGenerateAccountKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/accounts/admin", path.GenerateAccountKey("admin"))
}

func TestGenerateAccountSecretKey(t *testing.T) {
	assert.Equal(t, "/cse-sr/rbac/secret", path.GenerateRBACSecretKey())
}

func TestGenerateDependencyRuleKey(t *testing.T) {
	// consumer
	k := path.GenerateConsumerDependencyRuleKey("a", nil)
	if k != "/cse-sr/ms/dep-rules/a/c" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
	k = path.GenerateConsumerDependencyRuleKey("a", &discovery.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "3",
		Version:     "4",
	})
	assert.Equal(t, "/cse-sr/ms/dep-rules/a/c/1/2/3/4", k)

	// provider
	k = path.GenerateProviderDependencyRuleKey("a", nil)
	assert.Equal(t, "/cse-sr/ms/dep-rules/a/p", k)

	k = path.GenerateProviderDependencyRuleKey("a", &discovery.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "3",
		Version:     "4",
	})
	assert.Equal(t, "/cse-sr/ms/dep-rules/a/p/1/2/3/4", k)
}
