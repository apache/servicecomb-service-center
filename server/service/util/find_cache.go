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
	"errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/karlseguin/ccache"
)

var FindInstancesCache = &VersionRuleCache{
	c: ccache.Layered(ccache.Configure()),
}

type VersionRuleCacheItem struct {
	Instances []*pb.MicroServiceInstance
	Rev       string
}

type VersionRuleCache struct {
	c *ccache.LayeredCache
}

func (c *VersionRuleCache) primaryKey(domainProject string, provider *pb.MicroServiceKey) string {
	return util.StringJoin([]string{
		domainProject,
		provider.Environment,
		provider.AppId,
		provider.ServiceName}, "/")
}

func (c *VersionRuleCache) Get(domainProject, consumer string, provider *pb.MicroServiceKey) *VersionRuleCacheItem {
	item, _ := c.c.Fetch(c.primaryKey(domainProject, provider), provider.Version, cacheTTL, func() (interface{}, error) {
		return nil, errors.New("not exist")
	})
	if item == nil || item.Expired() {
		return nil
	}

	if v, ok := item.Value().(*util.ConcurrentMap).Get(consumer); ok {
		return v.(*VersionRuleCacheItem)
	}
	return nil
}

func (c *VersionRuleCache) Set(domainProject, consumer string, provider *pb.MicroServiceKey, item *VersionRuleCacheItem) {
	c2, _ := c.c.Fetch(c.primaryKey(domainProject, provider), provider.Version, cacheTTL, func() (interface{}, error) {
		// new one if not exist
		return util.NewConcurrentMap(1), nil
	})
	c2.Value().(*util.ConcurrentMap).Put(consumer, item)
}

func (c *VersionRuleCache) Delete(domainProject, consumer string, provider *pb.MicroServiceKey) {
	c.c.DeleteAll(c.primaryKey(domainProject, provider))
}
