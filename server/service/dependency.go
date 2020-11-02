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

package service

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

func (s *MicroServiceService) AddDependenciesForMicroServices(ctx context.Context, in *pb.AddDependenciesRequest) (*pb.AddDependenciesResponse, error) {
	if err := Validate(in); err != nil {
		return &pb.AddDependenciesResponse{
			Response: serviceUtil.BadParamsResponse(err.Error()).Response,
		}, nil
	}

	resp, err := datasource.Instance().AddOrUpdateDependencies(ctx, in.Dependencies, false)
	return &pb.AddDependenciesResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) CreateDependenciesForMicroServices(ctx context.Context, in *pb.CreateDependenciesRequest) (*pb.CreateDependenciesResponse, error) {
	if err := Validate(in); err != nil {
		return &pb.CreateDependenciesResponse{
			Response: serviceUtil.BadParamsResponse(err.Error()).Response,
		}, nil
	}

	resp, err := datasource.Instance().AddOrUpdateDependencies(ctx, in.Dependencies, true)
	return &pb.CreateDependenciesResponse{
		Response: resp,
	}, err
}

func (s *MicroServiceService) GetProviderDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "GetProviderDependencies failed for validating parameters failed")
		return &pb.GetProDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().SearchProviderDependency(ctx, in)
}

func (s *MicroServiceService) GetConsumerDependencies(ctx context.Context, in *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error) {
	err := Validate(in)
	if err != nil {
		log.Errorf(err, "GetConsumerDependencies failed for validating parameters failed")
		return &pb.GetConDependenciesResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, nil
	}

	return datasource.Instance().SearchConsumerDependency(ctx, in)
}
