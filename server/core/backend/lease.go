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
package backend

import (
	errorsEx "github.com/apache/incubator-servicecomb-service-center/pkg/errors"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	"golang.org/x/net/context"
	"time"
)

type LeaseTask struct {
	key        string
	LeaseID    int64
	TTL        int64
	CreateTime time.Time
	StartTime  time.Time
	EndTime    time.Time
	err        error
}

func (lat *LeaseTask) Key() string {
	return lat.key
}

func (lat *LeaseTask) Do(ctx context.Context) (err error) {
	lat.StartTime = time.Now()
	lat.TTL, err = Registry().LeaseRenew(ctx, lat.LeaseID)
	lat.EndTime = time.Now()
	if err != nil {
		util.Logger().Errorf(err, "[%s]renew lease %d failed(rev: %s, run: %s), key %s",
			time.Now().Sub(lat.CreateTime),
			lat.LeaseID,
			lat.CreateTime.Format(TIME_FORMAT),
			lat.StartTime.Format(TIME_FORMAT),
			lat.Key())
		if _, ok := err.(errorsEx.InternalError); !ok {
			// it means lease not found if err is not the InternalError type
			lat.err = err
			return
		}
	}

	lat.err, err = nil, nil
	util.LogNilOrWarnf(lat.CreateTime, "renew lease %d(rev: %s, run: %s), key %s",
		lat.LeaseID,
		lat.CreateTime.Format(TIME_FORMAT),
		lat.StartTime.Format(TIME_FORMAT),
		lat.Key())
	return
}

func (lat *LeaseTask) Err() error {
	return lat.err
}

func NewLeaseAsyncTask(op registry.PluginOp) *LeaseTask {
	return &LeaseTask{
		key:        ToLeaseAsyncTaskKey(util.BytesToStringWithNoCopy(op.Key)),
		LeaseID:    op.Lease,
		CreateTime: time.Now(),
	}
}

func ToLeaseAsyncTaskKey(key string) string {
	return "LeaseAsyncTask_" + key
}
