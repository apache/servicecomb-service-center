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

func PutManyTags(ctx context.Context, in *pb.AddServiceTagsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateAddServiceTagsRequest(in); err != nil {
		log.Error(fmt.Sprintf("add service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Tags, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().PutManyTags(ctx, in)
}

func PutTag(ctx context.Context, in *pb.UpdateServiceTagRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUpdateServiceTagRequest(in); err != nil {
		tagFlag := util.StringJoin([]string{in.Key, in.Value}, "/")
		log.Error(fmt.Sprintf("update service[%s]'s tag[%s] failed, operator: %s", in.ServiceId, tagFlag, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().PutTag(ctx, in)
}

func DeleteManyTags(ctx context.Context, in *pb.DeleteServiceTagsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateDeleteServiceTagsRequest(in); err != nil {
		log.Error(fmt.Sprintf("delete service[%s]'s tags %v failed, operator: %s", in.ServiceId, in.Keys, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().DeleteManyTags(ctx, in)
}

func ListTag(ctx context.Context, in *pb.GetServiceTagsRequest) (*pb.GetServiceTagsResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetServiceTagsRequest(in); err != nil {
		log.Error(fmt.Sprintf("get service[%s]'s tags failed, operator: %s", in.ServiceId, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().ListTag(ctx, in)
}
