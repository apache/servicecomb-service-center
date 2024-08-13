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

package server

import (
	"github.com/go-chassis/go-chassis/v2"
	chassisServer "github.com/go-chassis/go-chassis/v2/core/server"

	"github.com/apache/servicecomb-service-center/pkg/log"
	syncv1 "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/metrics"
	"github.com/apache/servicecomb-service-center/syncer/rpc"
	"github.com/apache/servicecomb-service-center/syncer/service/admin"
	"github.com/apache/servicecomb-service-center/syncer/service/sync"
)

// Run register chassis schema and run syncer services before chassis.Run()
func Run() {
	if err := config.Init(); err != nil {
		log.Error("syncer config init failed", err)
	}

	if !config.GetConfig().Sync.EnableOnStart {
		log.Warn("syncer is disabled")
		return
	}

	if len(config.GetConfig().Sync.Peers) <= 0 {
		log.Warn("peers parameter configuration is empty")
		return
	}

	if config.GetConfig().Sync.RbacEnabled {
		log.Info("syncer rbac enabled")
	}

	chassis.RegisterSchema("grpc", rpc.NewServer(),
		chassisServer.WithRPCServiceDesc(&syncv1.EventService_ServiceDesc))

	admin.Init()

	sync.Init()

	if err := metrics.Init(); err != nil {
		log.Error("syncer metrics init failed", err)
	}
}
