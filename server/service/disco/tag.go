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

package disco

import (
	"context"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/validator"
)

func (s *MicroServiceService) AddTags(ctx context.Context, in *pb.AddServiceTagsRequest) (*pb.AddServiceTagsResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Tags, remoteIP), err)
		return &pb.AddServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().AddTags(ctx, in)
}

func (s *MicroServiceService) UpdateTag(ctx context.Context, in *pb.UpdateServiceTagRequest) (*pb.UpdateServiceTagResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		tagFlag := util.StringJoin([]string{in.Key, in.Value}, "/")
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, operator: %s", in.ServiceId, tagFlag, remoteIP), err)
		return &pb.UpdateServiceTagResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().UpdateTag(ctx, in)
}

func (s *MicroServiceService) DeleteTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) (*pb.DeleteServiceTagsResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		remoteIP := util.GetIPFromContext(ctx)
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Keys, remoteIP), err)
		return &pb.DeleteServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().DeleteTags(ctx, in)
}

func (s *MicroServiceService) GetTags(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed", in.ServiceId), err)
		return &pb.GetServiceTagsResponse{
			Response: pb.CreateResponse(pb.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.GetMetadataManager().GetTags(ctx, in)
}
