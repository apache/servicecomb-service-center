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

package client

import (
	"context"
	"time"

	errorsEx "github.com/apache/servicecomb-service-center/pkg/errors"
	"github.com/apache/servicecomb-service-center/pkg/log"
	simple "github.com/apache/servicecomb-service-center/pkg/time"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

const leaseProfTimeFmt = "15:04:05.000"

type LeaseTask struct {
	Client Registry

	key     string
	LeaseID int64
	TTL     int64

	recvTime simple.Time
	err      error
}

func (lat *LeaseTask) Key() string {
	return lat.key
}

func (lat *LeaseTask) Do(ctx context.Context) (err error) {
	recv, start := lat.ReceiveTime(), time.Now()
	lat.TTL, err = lat.Client.LeaseRenew(ctx, lat.LeaseID)
	metrics.ReportHeartbeatCompleted(err, recv)
	if err != nil {
		log.Errorf(err, "[%s]task[%s] renew lease[%d] failed(recv: %s, send: %s)",
			time.Since(recv),
			lat.Key(),
			lat.LeaseID,
			recv.Format(leaseProfTimeFmt),
			start.Format(leaseProfTimeFmt))
		if _, ok := err.(errorsEx.InternalError); !ok {
			// it means lease not found if err is not the InternalError type
			lat.err = err
			return
		}
	}

	lat.err, err = nil, nil

	cost := time.Since(recv)
	if cost >= 2*time.Second {
		log.Warnf("[%s]task[%s] renew lease[%d](recv: %s, send: %s)",
			cost,
			lat.Key(),
			lat.LeaseID,
			recv.Format(leaseProfTimeFmt),
			start.Format(leaseProfTimeFmt))
	}
	return
}

func (lat *LeaseTask) Err() error {
	return lat.err
}

func (lat *LeaseTask) ReceiveTime() time.Time {
	return lat.recvTime.Local()
}

func NewLeaseAsyncTask(op PluginOp) *LeaseTask {
	return &LeaseTask{
		Client:   Instance(),
		key:      ToLeaseAsyncTaskKey(util.BytesToStringWithNoCopy(op.Key)),
		LeaseID:  op.Lease,
		recvTime: simple.FromTime(time.Now()),
	}
}

func ToLeaseAsyncTaskKey(key string) string {
	return "LeaseAsyncTask_" + key
}
