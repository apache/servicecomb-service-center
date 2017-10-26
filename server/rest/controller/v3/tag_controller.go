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
package v3

import (
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/rest/controller/v4"
)

type TagService struct {
	v4.TagService
}

func (this *TagService) URLPatterns() []rest.Route {
	return []rest.Route{
		{rest.HTTP_METHOD_POST, "/registry/v3/microservices/:serviceId/tags", this.AddTags},
		{rest.HTTP_METHOD_PUT, "/registry/v3/microservices/:serviceId/tags/:key", this.UpdateTag},
		{rest.HTTP_METHOD_GET, "/registry/v3/microservices/:serviceId/tags", this.GetTags},
		{rest.HTTP_METHOD_DELETE, "/registry/v3/microservices/:serviceId/tags/:key", this.DeleteTags},
	}
}
