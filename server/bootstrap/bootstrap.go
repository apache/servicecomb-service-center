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
package bootstrap

import _ "github.com/servicecomb/service-center/security"
import _ "github.com/servicecomb/service-center/common/logrotate"
import _ "github.com/servicecomb/service-center/security/plugins/plain"
import _ "github.com/servicecomb/service-center/server/core/registry/embededetcd"
import _ "github.com/servicecomb/service-center/server/core/registry/etcd"
import _ "github.com/servicecomb/service-center/server/plugins/infra/quota/buildin"
import _ "github.com/servicecomb/service-center/server/plugins/infra/quota/unlimit"
import "github.com/servicecomb/service-center/server/interceptor"
import "github.com/servicecomb/service-center/util"
import "github.com/servicecomb/service-center/server/interceptor/domain"
import "github.com/servicecomb/service-center/server/interceptor/maxbody"
import "github.com/servicecomb/service-center/server/interceptor/ratelimiter"

func init() {
	util.LOGGER.Info("BootStrap Huawei Enterprise Edition")

	interceptor.InterceptFunc(interceptor.ACCESS_PHASE, domain.Intercept)
	interceptor.InterceptFunc(interceptor.ACCESS_PHASE, ratelimiter.Intercept)

	interceptor.InterceptFunc(interceptor.CONTENT_PHASE, maxbody.Intercept)
}
