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
	"github.com/apache/servicecomb-service-center/server/config"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/go-chassis/cari/dlock"
	"github.com/robfig/cron/v3"
)

const (
	defaultRetireMicroserviceCron = "0 1 * * *"
	defaultReserveVersionCount    = 3
	retireServiceLockTTL          = 60
	retireServiceLockKey          = "retire-service-job"
)

func init() {
	disable := config.GetBool("registry.service.retire.disable", false)
	if disable {
		return
	}
	startRetireServiceJob()
}

func startRetireServiceJob() {
	localPlan := &datasource.RetirePlan{
		Cron:    config.GetString("registry.service.retire.cron", defaultRetireMicroserviceCron),
		Reserve: config.GetInt("registry.service.retire.reserve", defaultReserveVersionCount),
	}
	log.Info(fmt.Sprintf("start retire microservice job, plan is %v", localPlan))

	c := cron.New()
	_, err := c.AddFunc(localPlan.Cron, func() {
		retireService(localPlan)
	})
	if err != nil {
		log.Error("cron add func failed", err)
		return
	}
	c.Start()
}

func retireService(localPlan *datasource.RetirePlan) {
	if err := dlock.TryLock(retireServiceLockKey, retireServiceLockTTL); err != nil {
		log.Error(fmt.Sprintf("try lock %s failed", retireServiceLockKey), err)
		return
	}
	defer func() {
		if err := dlock.Unlock(retireServiceLockKey); err != nil {
			log.Error("unlock failed", err)
		}
	}()

	log.Info("start retire microservice")
	err := discosvc.RetireService(context.Background(), localPlan)
	if err != nil {
		log.Error("retire microservice failed", err)
	}
}
