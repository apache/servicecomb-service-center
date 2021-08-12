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

	"github.com/apache/servicecomb-service-center/pkg/log"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/go-chassis/foundation/gopool"
)

const (
	unusedMicroserviceCleanupInterval = 12 * time.Hour
	reserveVersionCount               = 3
)

func init() {
	startCleanupUnusedMicroserviceJob()
}

func startCleanupUnusedMicroserviceJob() {
	log.Info(fmt.Sprintf("start unused microservice cleanup job(every %s)", unusedMicroserviceCleanupInterval))
	gopool.Go(func(ctx context.Context) {
		tick := time.NewTicker(unusedMicroserviceCleanupInterval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				err := discosvc.CleanupUnusedMicroservice(ctx, reserveVersionCount)
				if err != nil {
					log.Error("cleanup lock history failed", err)
				}
			}
		}
	})
}
