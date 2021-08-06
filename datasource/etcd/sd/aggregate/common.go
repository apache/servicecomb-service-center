// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggregate

import (
	"fmt"
	"strings"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
)

const (
	Aggregate      = "aggregate"
	AggregateModes = "k8s,servicecenter"
)

var (
	closedCh      = make(chan struct{})
	repos         []string
	registryIndex = 0
)

func init() {
	close(closedCh)

	if Aggregate != config.GetString("discovery.kind", "", config.WithStandby("discovery_plugin")) {
		return
	}

	modes := config.GetString("discovery.aggregate.mode", AggregateModes, config.WithStandby("aggregate_mode"))
	repos = strings.Split(modes, ",")
	log.Info(fmt.Sprintf("aggregate_mode is %s", repos))

	// here save the index if found the registry plugin in modes list,
	// it is used for getting the one writable registry to handle requests
	// from API layer.
	registry := config.GetString("registry.kind", "", config.WithStandby("registry_plugin"))
	for i, repo := range repos {
		if repo == registry {
			registryIndex = i
			log.Info(fmt.Sprintf("found the registry index is %d", registryIndex))
			break
		}
	}
}
