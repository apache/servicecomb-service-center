// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"testing"
)

func TestGenerateDependencyRuleKey(t *testing.T) {
	// consumer
	k := GenerateConsumerDependencyRuleKey("a", nil)
	if k != "/cse-sr/ms/dep-rules/a/c" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
	k = GenerateConsumerDependencyRuleKey("a", &proto.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "3",
		Version:     "4",
	})
	if k != "/cse-sr/ms/dep-rules/a/c/1/2/3/4" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
	k = GenerateConsumerDependencyRuleKey("a", &proto.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "*",
		Version:     "4",
	})
	if k != "/cse-sr/ms/dep-rules/a/c/1/*" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}

	// provider
	k = GenerateProviderDependencyRuleKey("a", nil)
	if k != "/cse-sr/ms/dep-rules/a/p" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
	k = GenerateProviderDependencyRuleKey("a", &proto.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "3",
		Version:     "4",
	})
	if k != "/cse-sr/ms/dep-rules/a/p/1/2/3/4" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
	k = GenerateProviderDependencyRuleKey("a", &proto.MicroServiceKey{
		Environment: "1",
		AppId:       "2",
		ServiceName: "*",
		Version:     "4",
	})
	if k != "/cse-sr/ms/dep-rules/a/p/1/*" {
		t.Fatalf("TestGenerateDependencyRuleKey failed")
	}
}
