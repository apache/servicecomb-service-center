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
package service_test

import (
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server"
	"github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	_ "github.com/ServiceComb/service-center/server/plugin/infra/quota/buildin"
	_ "github.com/ServiceComb/service-center/server/plugin/infra/registry/etcd"
	_ "github.com/ServiceComb/service-center/server/plugin/infra/uuid/dynamic"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	"testing"
)

func TestGrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "model Suite", []Reporter{junitReporter})
}

var serviceResource pb.ServiceCtrlServer
var instanceResource pb.SerivceInstanceCtrlServerEx
var governService pb.GovernServiceCtrlServerEx

var _ = BeforeSuite(func() {
	//init plugin
	server.InitAPI()
	serviceResource = core.ServiceAPI
	instanceResource = core.InstanceAPI
	governService = core.GovernServiceAPI
})

func getContext() context.Context {
	ctx := context.TODO()
	ctx = util.SetContext(ctx, "domain", "default")
	ctx = util.SetContext(ctx, "project", "default")
	ctx = util.SetContext(ctx, "noCache", "1")
	return ctx
}
