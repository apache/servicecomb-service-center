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
package util

import (
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"testing"
)

func TestVersionRuleCache_Get(t *testing.T) {
	p := &pb.MicroServiceKey{}
	c := FindInstancesCache.Get("d", "c", p)
	if c != nil {
		t.Fatalf("TestVersionRuleCache_Get failed, %v", c)
	}

	r := &VersionRuleCacheItem{Rev: 1}
	FindInstancesCache.Set("d", "c", p, r)
	c = FindInstancesCache.Get("d", "c", p)
	if c == nil {
		t.Fatalf("TestVersionRuleCache_Get failed, %v", c)
	}
	if c.Rev != 1 {
		t.Fatalf("TestVersionRuleCache_Get failed, rev %d != 1", c.Rev)
	}
	c = FindInstancesCache.Get("d", "c2", p)
	if c != nil {
		t.Fatalf("TestVersionRuleCache_Get failed, %v", c)
	}

	p2 := &pb.MicroServiceKey{ServiceName: "p2"}
	FindInstancesCache.Set("d", "c2", p, r)
	FindInstancesCache.Set("d", "c2", p2, r)
	FindInstancesCache.Delete("d", "c", p)
	c = FindInstancesCache.Get("d", "c2", p)
	if c != nil {
		t.Fatalf("TestVersionRuleCache_Get failed, %v", c)
	}
	c = FindInstancesCache.Get("d", "c2", p2)
	if c == nil {
		t.Fatalf("TestVersionRuleCache_Get failed, %v", c)
	}
	if c.Rev != 1 {
		t.Fatalf("TestVersionRuleCache_Get failed, rev %d != 1", c.Rev)
	}
}
