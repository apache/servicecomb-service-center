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

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

type MicroServiceService struct {
}

func AddDependencies(ctx context.Context, in *pb.AddDependenciesRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateAddDependenciesRequest(in); err != nil {
		log.Error(fmt.Sprintf("AddDependencies failed, operator: %s", remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetDependencyManager().PutDependencies(ctx, in.Dependencies, false)
}

func PutDependencies(ctx context.Context, in *pb.CreateDependenciesRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateCreateDependenciesRequest(in); err != nil {
		log.Error(fmt.Sprintf("PutDependencies failed, operator: %s", remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetDependencyManager().PutDependencies(ctx, in.Dependencies, true)
}

func ListConsumers(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetDependenciesRequest(in); err != nil {
		log.Error(fmt.Sprintf("ListConsumers failed for validating parameters failed, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetDependencyManager().ListConsumers(ctx, in)
}

func ListProviders(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetDependenciesRequest(in); err != nil {
		log.Error(fmt.Sprintf("ListProviders failed for validating parameters failed, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetDependencyManager().ListProviders(ctx, in)
}
