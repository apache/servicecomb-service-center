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

package dump

import (
	"github.com/apache/servicecomb-service-center/pkg/cluster"
	"github.com/apache/servicecomb-service-center/server/alarm/model"
	"github.com/go-chassis/cari/discovery"
)

type AlarmListRequest struct {
}

type AlarmListResponse struct {
	Response *discovery.Response `json:"-"`
	Alarms   []*model.AlarmEvent `json:"alarms,omitempty"`
}

type ClustersRequest struct {
}

type ClustersResponse struct {
	Response *discovery.Response `json:"-"`
	Clusters cluster.Clusters    `json:"clusters,omitempty"`
}

type ClearAlarmRequest struct {
}

type ClearAlarmResponse struct {
	Response *discovery.Response `json:"-"`
}
