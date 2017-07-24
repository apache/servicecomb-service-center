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

import "errors"

type Worker interface {
	Err() error
	SetError(err error)
	Id() string
	Subject() string
	Service() *NotifyService
	SetService(*NotifyService)
	OnAccept()
	OnMessage(job NotifyJob)
	Close()
}

type NotifyJob interface {
	Id() string
	Subject() string
}

type BaseWorker struct {
	id      string
	subject string
	service *NotifyService
	err     error
}

func (s *BaseWorker) Err() error {
	return s.err
}

func (s *BaseWorker) SetError(err error) {
	s.err = err
	// 触发清理job
	s.Service().AddJob(&NotifyServiceHealthCheckJob{
		BaseNotifyJob: BaseNotifyJob{
			id:      NOTIFY_SERVER_CHECKER_NAME,
			subject: NOTIFY_SERVER_CHECK_SUBJECT,
		},
		ErrorWorker: s,
	})
}

func (s *BaseWorker) Id() string {
	return s.id
}

func (s *BaseWorker) Subject() string {
	return s.subject
}

func (s *BaseWorker) Service() *NotifyService {
	return s.service
}

func (s *BaseWorker) SetService(svc *NotifyService) {
	s.service = svc
}

func (s *BaseWorker) OnAccept() {
}

func (s *BaseWorker) OnMessage(job NotifyJob) {
	s.SetError(errors.New("do not call base notifier OnMessage method"))
}

func (s *BaseWorker) Close() {

}

type BaseNotifyJob struct {
	id      string
	subject string
}

func (s *BaseNotifyJob) Id() string {
	return s.id
}

func (s *BaseNotifyJob) Subject() string {
	return s.subject
}
