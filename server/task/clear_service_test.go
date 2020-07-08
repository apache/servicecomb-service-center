// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task_test

// initialize
import _ "github.com/apache/servicecomb-service-center/server/bootstrap"
import _ "github.com/apache/servicecomb-service-center/server"

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/server/task"
	"github.com/astaxie/beego"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	testing.Init()
	apt.Initialize()
}

var _ = BeforeSuite(func() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	//clear service created in last test
	time.Sleep(timeLimit)
	task.ClearNoInstanceServices(timeLimit)
})

// map[domainProject][serviceName]*serviceCleanInfo
var svcCleanInfos = make(map[string]map[string]*serviceCleanInfo)
var timeLimit = 2 * time.Second

type serviceCleanInfo struct {
	ServiceName  string
	ServiceId    string
	WithInstance bool
	ShouldClear  bool
}

func TestTask(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Task Suite")
}

func getContext(domain string, project string) context.Context {
	return util.SetContext(
		util.SetDomainProject(context.Background(), domain, project),
		util.CtxNocache, "1")
}

func createService(domain string, project string, name string, withInstance bool, shouldClear bool) {
	By(fmt.Sprintf("create service: %s, with instance: %t, should clear: %t", name, withInstance, shouldClear))
	svc := &pb.CreateServiceRequest{
		Service: &pb.MicroService{
			AppId:       "clear",
			ServiceName: name,
			Version:     "1.0",
		},
	}
	if withInstance {
		svc.Instances = []*pb.MicroServiceInstance{
			{
				Endpoints: []string{"http://127.0.0.1:80"},
				HostName:  "1",
			},
		}
	}
	ctx := getContext(domain, project)
	svcResp, err := apt.ServiceAPI.Create(ctx, svc)
	Expect(err).To(BeNil())
	Expect(svcResp).NotTo(BeNil())
	Expect(svcResp.Response.GetCode()).To(Equal(pb.Response_SUCCESS))
	info := &serviceCleanInfo{
		ServiceName:  name,
		ServiceId:    svcResp.ServiceId,
		WithInstance: withInstance,
		ShouldClear:  shouldClear,
	}
	domainProject := domain + apt.SPLIT + project
	m, ok := svcCleanInfos[domainProject]
	if !ok {
		m = make(map[string]*serviceCleanInfo)
		svcCleanInfos[domainProject] = m
	}
	m[name] = info
}

func checkServiceCleared(domain string, project string) {
	domainProject := domain + apt.SPLIT + project
	m := svcCleanInfos[domainProject]
	for _, v := range m {
		By(fmt.Sprintf("check cleared, service: %s, should be cleared: %t", v.ServiceName, v.ShouldClear))
		getSvcReq := &pb.GetServiceRequest{
			ServiceId: v.ServiceId,
		}
		ctx := getContext(domain, project)
		getSvcResp, err := apt.ServiceAPI.GetOne(ctx, getSvcReq)
		Expect(err).To(BeNil())
		Expect(getSvcResp).NotTo(BeNil())
		Expect(getSvcResp.Response.GetCode() == pb.Response_SUCCESS).To(Equal(!v.ShouldClear))
	}
}

func serviceClearCheckFunc(domain string, project string) func() {
	return func() {
		var err error
		It("should run clear task success", func() {
			withInstance := true
			withNoInstance := false
			shouldClear := true
			shouldNotClear := false

			createService(domain, project, "svc1", withNoInstance, shouldClear)
			createService(domain, project, "svc2", withInstance, shouldNotClear)
			time.Sleep(timeLimit)
			createService(domain, project, "svc3", withNoInstance, shouldNotClear)
			createService(domain, project, "svc4", withInstance, shouldNotClear)

			err = task.ClearNoInstanceServices(timeLimit)
			Expect(err).To(BeNil())

			checkServiceCleared(domain, project)
		})
	}
}

var _ = Describe("clear service", func() {
	Describe("domain project 1", serviceClearCheckFunc("default1", "default"))
	Describe("domain project 2", serviceClearCheckFunc("default2", "default"))
})
var _ = func() bool {

	return true
}()
