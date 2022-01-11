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

package govern

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

func ListServiceDetail(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	ctx = util.WithCacheOnly(ctx)
	return datasource.GetMetadataManager().ListServiceDetail(ctx, in)
}

func GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.ServiceDetail, error) {
	ctx = util.WithCacheOnly(ctx)

	if len(in.ServiceId) == 0 {
		return nil, pb.NewError(pb.ErrInvalidParams, "Invalid request for getting service detail.")
	}

	return datasource.GetMetadataManager().GetServiceDetail(ctx, in)
}

func ListApp(ctx context.Context, in *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	if err := validator.ValidateGetAppsRequest(in); err != nil {
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().ListApp(ctx, in)
}

func GetOverview(ctx context.Context, in *pb.GetServicesRequest) (*pb.Statistics, error) {
	ctx = util.WithCacheOnly(ctx)
	return datasource.GetMetadataManager().GetOverview(ctx, in)
}
