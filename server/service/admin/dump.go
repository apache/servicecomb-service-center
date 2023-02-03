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

package admin

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/version"
	mapset "github.com/deckarep/golang-set"
	"github.com/go-chassis/cari/discovery"
)

func Dump(ctx context.Context, in *dump.Request) (*dump.Response, error) {
	domainProject := util.ParseDomainProject(ctx)
	if !datasource.IsDefaultDomainProject(domainProject) {
		return nil, discovery.NewError(discovery.ErrForbidden, "Required admin permission")
	}
	resp := &dump.Response{}
	set := toSet(in.Options)
	if set.Cardinality() == 0 {
		appendData(ctx, "cache", resp)
		return resp, nil
	}
	set.Each(func(option interface{}) bool {
		appendData(ctx, option.(string), resp)
		return true
	})
	return resp, nil
}

func toSet(arr []string) mapset.Set {
	if len(arr) == 0 {
		return mapset.NewSet()
	}
	set := mapset.NewSet()
	for _, kind := range arr {
		if kind == "all" {
			return mapset.NewSet("all")
		}
		set.Add(kind)
	}
	return set
}

func appendData(ctx context.Context, option string, resp *dump.Response) {
	switch option {
	case "info":
		resp.Info = version.Ver()
	case "cache":
		resp.Cache = datasource.GetSystemManager().DumpCache(ctx)
	case "all":
		appendData(ctx, "info", resp)
		appendData(ctx, "cache", resp)
	}
}
