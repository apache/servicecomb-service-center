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
	roa "github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rpc"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"google.golang.org/grpc"
)

func init() {
	registerGRPC()

	registerREST()
}

func registerGRPC() {
	rpc.RegisterService(func(s *grpc.Server) {
		pb.RegisterGovernServiceCtrlServer(s, GovernServiceAPI)
	})
}

func registerREST() {
	roa.RegisterServant(&GovernServiceControllerV3{})
	roa.RegisterServant(&GovernServiceControllerV4{})
}
