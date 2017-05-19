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

type Notifier interface {
        Err() error
        SetError(err error)
        GetId() string
        GetSubject() string
        GetServer() *NotifyService
        Notify(job NotifyJob)
        Close()
}

type NotifyJob interface {
        GetId() string
        GetSubject() string
}

type BaseNotifier struct {
        Id      string
        Subject string
        Server  *NotifyService
        err     error
}

func (s *BaseNotifier) Err() error {
        return s.err
}

func (s *BaseNotifier) SetError(err error) {
        s.err = err
        // 触发清理job
        s.Server.AddJob(&NotifyServiceHealthCheckJob{
                BaseNotifyJob: BaseNotifyJob{
                        Id:      NOTIFY_SERVER_CHECKER_NAME,
                        Subject: NOTIFY_SERVER_CHECK_SUBJECT,
                },
                ErrorNotifier: s,
        })
}

func (s *BaseNotifier) GetId() string {
        return s.Id
}

func (s *BaseNotifier) GetSubject() string {
        return s.Subject
}

func (s *BaseNotifier) GetServer() *NotifyService {
        return s.Server
}

func (s *BaseNotifier) Notify(job NotifyJob) {
        s.SetError(errors.New("do not call base notifier notify method"))
}

func (s *BaseNotifier) Close() {

}

type BaseNotifyJob struct {
        Id      string
        Subject string
}

func (s *BaseNotifyJob) GetId() string {
        return s.Id
}

func (s *BaseNotifyJob) GetSubject() string {
        return s.Subject
}
