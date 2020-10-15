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

package datasource

import (
	"context"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
)

type Dependency interface {
	AddDependency(ctx context.Context, request *pb.AddDependenciesRequest) (*pb.AddDependenciesResponse, error)
	CreateDependency(ctx context.Context, request *pb.CreateDependenciesRequest) (*pb.CreateDependenciesResponse, error)
	SearchProviderDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetProDependenciesResponse, error)
	SearchConsumerDependency(ctx context.Context, request *pb.GetDependenciesRequest) (*pb.GetConDependenciesResponse, error)
	DeleteDependency()
}
