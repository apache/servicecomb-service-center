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

package tombstone

import (
	"context"

	"github.com/go-chassis/cari/sync"

	"github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/apache/servicecomb-service-center/eventbase/model"
)

func Get(ctx context.Context, req *model.GetTombstoneRequest) (*sync.Tombstone, error) {
	if len(req.Domain) == 0 {
		req.Domain = sync.Default
	}
	if len(req.Project) == 0 {
		req.Project = sync.Default
	}
	return datasource.GetTombstoneDao().Get(ctx, req)
}

func Delete(ctx context.Context, tombstones ...*sync.Tombstone) error {
	return datasource.GetTombstoneDao().Delete(ctx, tombstones...)
}

func List(ctx context.Context, request *model.ListTombstoneRequest) ([]*sync.Tombstone, error) {
	if len(request.Domain) == 0 {
		request.Domain = sync.Default
	}
	if len(request.Project) == 0 {
		request.Project = sync.Default
	}
	opts := []datasource.TombstoneFindOption{
		datasource.WithTombstoneDomain(request.Domain),
		datasource.WithTombstoneProject(request.Project),
		datasource.WithResourceType(request.ResourceType),
		datasource.WithBeforeTimestamp(request.BeforeTimestamp),
	}
	return datasource.GetTombstoneDao().List(ctx, opts...)
}
