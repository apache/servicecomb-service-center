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

package admin

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/alarm"
)

func Clusters(ctx context.Context, in *dump.ClustersRequest) (*dump.ClustersResponse, error) {
	clusters, err := datasource.GetSCManager().GetClusters(ctx)
	if err != nil {
		return nil, err
	}
	return &dump.ClustersResponse{
		Clusters: clusters,
	}, nil
}

func AlarmList(ctx context.Context, in *dump.AlarmListRequest) (*dump.AlarmListResponse, error) {
	return &dump.AlarmListResponse{
		Alarms: alarm.ListAll(),
	}, nil
}

func ClearAlarm(ctx context.Context, in *dump.ClearAlarmRequest) (*dump.ClearAlarmResponse, error) {
	alarm.ClearAll()
	log.Infof("service center alarms are cleared")
	return &dump.ClearAlarmResponse{}, nil
}
