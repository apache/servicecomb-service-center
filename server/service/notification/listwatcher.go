//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package notification

import (
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/util"
)

// 状态变化推送
type WatchJob struct {
	BaseNotifyJob
	Revision int64
	Response *pb.WatchInstanceResponse
}

type ListWatcher struct {
	BaseSubscriber
	Job          chan NotifyJob
	ListRevision int64
	ListFunc     func() (results []*pb.WatchInstanceResponse, rev int64)

	listCh chan struct{}
}

func (w *ListWatcher) OnAccept() {
	if w.Err() != nil {
		return
	}

	util.LOGGER.Debugf("accepted by notify service, current revision is %v", w.ListRevision)
	go w.listAndPublishJobs()
}

func (w *ListWatcher) listAndPublishJobs() {
	defer close(w.listCh)
	if w.ListFunc == nil {
		return
	}
	results, rev := w.ListFunc()
	w.ListRevision = rev
	for _, response := range results {
		w.sendMessage(NewWatchJob(w.Type(), w.Id(), w.Subject(), w.ListRevision, response))
	}
}

//被通知
func (w *ListWatcher) OnMessage(job NotifyJob) {
	if w.Err() != nil {
		return
	}

	select {
	case _, ok := <-w.listCh:
		if !ok {
			// means list operation is over
		}
	}

	if job.(*WatchJob).Revision <= w.ListRevision {
		util.LOGGER.Warnf(nil, "unexpected notify job is coming in, job is %v, current revision is %v",
			job, w.ListRevision)
		return
	}
	w.sendMessage(job)
}

func (w *ListWatcher) sendMessage(job NotifyJob) {
	util.LOGGER.Debugf("start notify watcher, job is %v, current revision is %v", job)
	defer util.RecoverAndReport()
	w.Job <- job
}

func (w *ListWatcher) Close() {
	close(w.Job)
}

func NewWatchJob(nType NotifyType, subscriberId, subject string, rev int64, response *pb.WatchInstanceResponse) *WatchJob {
	return &WatchJob{
		BaseNotifyJob: BaseNotifyJob{
			subscriberId: subscriberId,
			subject:      subject,
			nType:        nType,
		},
		Revision: rev,
		Response: response,
	}
}

func NewWatcher(nType NotifyType, id string, subject string) *ListWatcher {
	return NewListWatcher(nType, id, subject, nil)
}

func NewListWatcher(nType NotifyType, id string, subject string,
	listFunc func() (results []*pb.WatchInstanceResponse, rev int64)) *ListWatcher {
	watcher := &ListWatcher{
		BaseSubscriber: BaseSubscriber{
			id:      id,
			subject: subject,
			nType:   nType,
		},
		Job:      make(chan NotifyJob, DEFAULT_MAX_QUEUE),
		ListFunc: listFunc,
		listCh:   make(chan struct{}),
	}
	return watcher
}
