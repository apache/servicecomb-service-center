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

import "github.com/servicecomb/service-center/util"

//Notifier 健康检查
type NotifyServiceHealthChecker struct {
        BaseNotifier
}

type NotifyServiceHealthCheckJob struct {
        BaseNotifyJob
        ErrorNotifier Notifier
}

func (s *NotifyServiceHealthChecker) Notify(job NotifyJob) {
        j := job.(*NotifyServiceHealthCheckJob)
        err := j.ErrorNotifier.Err()
        util.LOGGER.Warnf(err, "notify server remove watcher %s %s, %s",
                j.ErrorNotifier.GetSubject(), j.ErrorNotifier.GetId(), err.Error())
        s.Server.RemoveNotifier(j.ErrorNotifier)
}
