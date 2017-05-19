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
package routers

import (
	"net/http"

	rs "github.com/servicecomb/service-center/server/rest"
	"github.com/servicecomb/service-center/util"
	roa "github.com/servicecomb/service-center/util/rest"
)

func initRouter() {
	roa.RegisterServent(&rs.MainService{})
	roa.RegisterServent(&rs.MicroServiceService{})
	roa.RegisterServent(&rs.TagService{})
	roa.RegisterServent(&rs.RuleService{})
	roa.RegisterServent(&rs.MicroServiceInstanceService{})
	roa.RegisterServent(&rs.WatchService{})
	roa.RegisterServent(&rs.GovernService{})
}

//GetRouter return the router fo REST service
func GetRouter() (router http.Handler) {
	util.LOGGER.Warn("init router", nil)
	router = roa.InitROAServerHandler()
	initRouter()

	return
}
