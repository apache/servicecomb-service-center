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
package notification

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"golang.org/x/net/context"
	"time"
)

// 状态变化推送
type WatchJob struct {
	*BaseNotifyJob
	Revision int64
	Response *pb.WatchInstanceResponse
}

type ListWatcher struct {
	*BaseSubscriber
	Job          chan *WatchJob
	ListRevision int64
	ListFunc     func() (results []*pb.WatchInstanceResponse, rev int64)

	listCh chan struct{}
}

func (s *ListWatcher) SetError(err error) {
	s.BaseSubscriber.SetError(err)
	// 触发清理job
	s.Service().AddJob(NewNotifyServiceHealthCheckJob(s))
}

func (w *ListWatcher) OnAccept() {
	if w.Err() != nil {
		return
	}

	util.Logger().Debugf("accepted by notify service, %s watcher %s %s", w.Type(), w.Group(), w.Subject())
	util.Go(w.listAndPublishJobs)
}

func (w *ListWatcher) listAndPublishJobs(_ context.Context) {
	defer close(w.listCh)
	if w.ListFunc == nil {
		return
	}
	results, rev := w.ListFunc()
	w.ListRevision = rev
	for _, response := range results {
		w.sendMessage(NewWatchJob(w.Group(), w.Subject(), w.ListRevision, response))
	}
}

//被通知
func (w *ListWatcher) OnMessage(job NotifyJob) {
	if w.Err() != nil {
		return
	}

	wJob, ok := job.(*WatchJob)
	if !ok {
		return
	}

	timer := time.NewTimer(GetNotifyService().Config.AddTimeout)
	select {
	case <-w.listCh:
		timer.Stop()
	case <-timer.C:
		util.Logger().Errorf(nil,
			"the %s listwatcher %s %s is not ready[over %s], send the event %v",
			w.Type(), w.Group(), w.Subject(), GetNotifyService().Config.AddTimeout, job)
	}

	if wJob.Revision <= w.ListRevision {
		util.Logger().Warnf(nil,
			"unexpected notify %s job is coming in, watcher %s %s, job is %v, current revision is %v",
			w.Type(), w.Group(), w.Subject(), job, w.ListRevision)
		return
	}
	w.sendMessage(wJob)
}

func (w *ListWatcher) sendMessage(job *WatchJob) {
	util.Logger().Debugf("start to notify %s watcher %s %s, job is %v, current revision is %v", w.Type(),
		w.Group(), w.Subject(), job, w.ListRevision)
	defer util.RecoverAndReport()
	timer := time.NewTimer(GetNotifyService().Config.AddTimeout)
	select {
	case w.Job <- job:
		timer.Stop()
	case <-timer.C:
		util.Logger().Errorf(nil,
			"the %s watcher %s %s event queue is full[over %s], drop the event %v",
			w.Type(), w.Group(), w.Subject(), GetNotifyService().Config.AddTimeout, job)
	}
}

func (w *ListWatcher) Close() {
	close(w.Job)
}

func NewWatchJob(subscriberId, subject string, rev int64, response *pb.WatchInstanceResponse) *WatchJob {
	return &WatchJob{
		BaseNotifyJob: &BaseNotifyJob{
			group:   subscriberId,
			subject: subject,
			nType:   INSTANCE,
		},
		Revision: rev,
		Response: response,
	}
}

func NewListWatcher(group string, subject string,
	listFunc func() (results []*pb.WatchInstanceResponse, rev int64)) *ListWatcher {
	watcher := &ListWatcher{
		BaseSubscriber: NewSubscriber(INSTANCE, subject, group),
		Job:            make(chan *WatchJob, GetNotifyService().Config.MaxQueue),
		ListFunc:       listFunc,
		listCh:         make(chan struct{}),
	}
	return watcher
}
