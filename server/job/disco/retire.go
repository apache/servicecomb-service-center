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
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/go-chassis/foundation/gopool"
)

const (
	defaultRetireMicroserviceInterval = 12 * time.Hour
	defaultReserveVersionCount        = 3
)

func init() {
	startRetireServiceJob()
}

func startRetireServiceJob() {
	disable := config.GetBool("registry.service.retire.disable", false)
	if disable {
		return
	}

	localPlan := &datasource.RetirePlan{
		Interval: config.GetDuration("registry.service.retire.interval", defaultRetireMicroserviceInterval),
		Reserve:  config.GetInt("registry.service.retire.reserve", defaultReserveVersionCount),
	}

	log.Info(fmt.Sprintf("start retire microservice job, plan is %v", localPlan))
	gopool.Go(func(ctx context.Context) {
		tick := time.NewTicker(localPlan.Interval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				err := discosvc.RetireService(ctx, localPlan)
				if err != nil {
					log.Error("retire microservice failed", err)
				}
			}
		}
	})
}
