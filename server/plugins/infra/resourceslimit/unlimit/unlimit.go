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
package unlimit

import (
	rl "github.com/ServiceComb/service-center/server/infra/resourceslimit"
)

type Unlimit struct {
}

func init() {
	rl.ResourcesLimitPlugins["unlimit"] = New
}

func (buildin *Unlimit) CanApplyResource(resourceType rl.ResourceType, tenant string, serviceId string, applyCount int32) (bool, error) {
	return true, nil
}

func New() rl.ResourcesManager {
	return &Unlimit{}
}
