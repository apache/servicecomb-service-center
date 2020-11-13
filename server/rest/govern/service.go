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
	"github.com/apache/servicecomb-service-center/pkg/proto"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	scerr "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/server/service"
)

var ServiceAPI proto.GovernServiceCtrlServer = &Service{}

type Service struct {
}

func (governService *Service) GetServicesInfo(ctx context.Context, in *pb.GetServicesInfoRequest) (*pb.GetServicesInfoResponse, error) {
	ctx = util.SetContext(ctx, util.CtxCacheOnly, "1")

	return datasource.Instance().GetServicesInfo(ctx, in)
}

func (governService *Service) GetServiceDetail(ctx context.Context, in *pb.GetServiceRequest) (*pb.GetServiceDetailResponse, error) {
	ctx = util.SetContext(ctx, util.CtxCacheOnly, "1")

	if len(in.ServiceId) == 0 {
		return &pb.GetServiceDetailResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Invalid request for getting service detail."),
		}, nil
	}

	return datasource.Instance().GetServiceDetail(ctx, in)
}

func (governService *Service) GetApplications(ctx context.Context, in *pb.GetAppsRequest) (*pb.GetAppsResponse, error) {
	err := service.Validate(in)
	if err != nil {
		return &pb.GetAppsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().GetApplications(ctx, in)
}
