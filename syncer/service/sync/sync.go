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

package sync

import (
	"fmt"

	// glint
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/event"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator"
	"github.com/apache/servicecomb-service-center/syncer/service/task"
)

func Init() {
	if !config.GetConfig().Sync.EnableOnStart {
		log.Info("sync not enabled")
		return
	}

	Work()
}

func Work() {
	work()
}

func work() {
	err := replicator.Work()
	if err != nil {
		log.Warn(fmt.Sprintf("replicate work init failed, %s", err.Error()))
		return
	}

	event.Work()

	task.Work()
}
