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
	"context"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	_ "github.com/ServiceComb/service-center/server/core/registry"
	_ "github.com/ServiceComb/service-center/server/core/registry/etcd"
	_ "github.com/ServiceComb/service-center/server/plugins/infra/quota/buildin"
	"github.com/ServiceComb/service-center/server/service"
	"testing"
)

const (
	TOO_LONG_SERVICEID   = "addasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasadafd"
	TOO_LONG_SERVICENAME = "addasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasaddasdfasadafd"
)

func TestGrpc(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "model Suite", []Reporter{junitReporter})
}

var serviceResource pb.ServiceCtrlServer
var insResource pb.SerivceInstanceCtrlServerEx
var governService pb.GovernServiceCtrlServerEx

var _ = BeforeSuite(func() {
	//init plugin
	serviceResource, insResource, governService = service.AssembleResources(nil)

})

func getContext() context.Context {
	ctx := context.TODO()
	ctx = context.WithValue(ctx, "tenant", "default")
	ctx = context.WithValue(ctx, "project", "default")
	return ctx
}
