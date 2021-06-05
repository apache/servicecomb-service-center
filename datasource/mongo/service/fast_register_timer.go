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

package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/apache/servicecomb-service-center/datasource/mongo/event"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	loopTime              = 100 * time.Millisecond
	batchLen              = 10000
	fuseMinCount          = 3
	fuseTime              = 5 * time.Second
	maxRegisterFailedTime = 500
	ctxCancelTimeOut      = 60 * time.Second
)

var fastRegisterTimeTask *FastRegisterTimeTask

type FastRegisterTimeTask struct {
	goroutine *gopool.Pool
}

func NewRegisterTimeTask() *FastRegisterTimeTask {
	return &FastRegisterTimeTask{
		goroutine: gopool.New(context.Background()),
	}
}

func (rt *FastRegisterTimeTask) Start() {
	gopool.Go(fastRegisterTimeTask.loopRegister)
}

func (rt *FastRegisterTimeTask) Stop() {
	rt.goroutine.Close(true)
}

func (rt *FastRegisterTimeTask) loopRegister(ctx context.Context) {
	blockCh := make(chan struct{}, runtime.NumCPU()-1)
	failedCount := make(chan int, 1)
	failedCount <- 0
	ticker := time.NewTicker(loopTime)

	for {
		select {
		case <-ctx.Done():
			// server shutdown
			return
		case <-ticker.C:
			length := len(GetFastRegisterInstanceService().InstEventCh)
			if length == 0 {
				continue
			}

			// fuse is triggered if failed registry counts more than fuseMinCount
			count := <-failedCount
			if count > fuseMinCount {
				time.Sleep(fuseTime)
			}
			failedCount <- count

			events := rt.generateEvents(length)
			blockCh <- struct{}{}
			go rt.RegisterInstancesAsync(events, blockCh, failedCount)

		case event, ok := <-GetFastRegisterInstanceService().FailedInstCh:
			// if instance batch register failed, register it single
			if !ok {
				log.Error("failed instance channel is error", errors.New("channel closed"))
				continue
			}

			if event.FailedTime > maxRegisterFailedTime {
				log.Error(fmt.Sprintf("instance register retry time is more than max register time:%d, "+
					"the instance params maybe wrong, drop it", maxRegisterFailedTime),
					errors.New("retry register instance failed"))
				continue
			}

			blockCh <- struct{}{}

			go rt.RegisterInstance(event, blockCh)
		}
	}
}

func (rt *FastRegisterTimeTask) generateEvents(length int) []*event.InstanceRegisterEvent {
	// if channel len >= batch len, use batch len, otherwise use channel len
	if length >= batchLen {
		length = batchLen
	}

	events := make([]*event.InstanceRegisterEvent, 0)

	for i := 0; i < length; i++ {
		event, ok := <-GetFastRegisterInstanceService().InstEventCh

		refreshCanceledCtx(event)

		if !ok {
			log.Error("instance event channel is error", errors.New("channel closed"))
			continue
		}
		events = append(events, event)
	}
	return events
}

func (rt *FastRegisterTimeTask) RegisterInstancesAsync(events []*event.InstanceRegisterEvent, blockCh chan struct{}, failedCount chan int) {
	defer endBlock(blockCh)

	ctx, cancel := context.WithTimeout(context.Background(), ctxCancelTimeOut)
	defer cancel()

	_, err := batchRegisterInstance(ctx, events)

	count := <-failedCount

	if err != nil {
		//failed count
		count = count + 1
		failedCount <- count

		//add to failed instance channel, will retry register
		log.Error("register instances err, retry it", err)
		GetFastRegisterInstanceService().AddFailedEvents(events)
		return
	}

	failedCount <- 0
}

func (rt *FastRegisterTimeTask) RegisterInstance(event *event.InstanceRegisterEvent, blockCh chan struct{}) {
	defer endBlock(blockCh)

	cancel := refreshCanceledCtx(event)
	defer cancel()

	_, err := registerInstanceSingle(event.Ctx, event.Request, event.IsCustomID)

	if err != nil {
		log.Error(fmt.Sprintf("register instance:%s failed again, failed times:%d",
			event.Request.Instance.InstanceId, event.FailedTime), err)
		event.FailedTime = event.FailedTime + 1
		GetFastRegisterInstanceService().AddFailedEvent(event)
	}
}

func endBlock(blockCh chan struct{}) {
	<-blockCh
}

func refreshCanceledCtx(event *event.InstanceRegisterEvent) context.CancelFunc {
	oldCtx := event.Ctx
	newCtx, cancel := context.WithTimeout(context.Background(), ctxCancelTimeOut)
	newCtx = util.SetDomain(newCtx, util.ParseDomain(oldCtx))
	newCtx = util.SetProject(newCtx, util.ParseProject(oldCtx))
	newCtx = util.SetContext(newCtx, util.CtxRemoteIP, util.ParseProject(oldCtx))
	event.Ctx = newCtx
	return cancel
}
