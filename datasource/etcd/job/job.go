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

package job

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"time"
)

// clear services who have no instance
func ClearNoInstanceServices() {
	if !config.ServerInfo.Config.ServiceClearEnabled {
		return
	}
	log.Infof("service clear enabled, interval: %s, service TTL: %s",
		config.ServerInfo.Config.ServiceClearInterval,
		config.ServerInfo.Config.ServiceTTL)

	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(config.ServerInfo.Config.ServiceClearInterval):
				lock, err := mux.Try(mux.ServiceClearLock)
				if err != nil {
					log.Errorf(err, "can not clear no instance services by this service center instance now")
					continue
				}
				err = datasource.Instance().ClearNoInstanceServices(ctx, config.ServerInfo.Config.ServiceTTL)
				if err := lock.Unlock(); err != nil {
					log.Error("", err)
				}
				if err != nil {
					log.Errorf(err, "no-instance services cleanup failed")
					continue
				}
				log.Info("no-instance services cleanup succeed")
			}
		}
	})
}
