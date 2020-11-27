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

package syncer

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/rest"
)

// Syncer 有关的接口
type SyncerController struct {
}

// URLPatterns 路由
func (ctrl *SyncerController) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/syncer/watch", Func: ctrl.WatchInstance},
	}
}

func (ctrl *SyncerController) WatchInstance(w http.ResponseWriter, r *http.Request) {
	syncerEnabled := config.GetBool("syncer.enabled", false)
	if syncerEnabled {
		ServiceAPI.WatchInstance(w, r)
	} else {
		log.Warnf("syncer cannot watch instance because the config syncer.enabled is false, ")
	}

}
