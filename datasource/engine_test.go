/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package datasource_test

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"

	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// map[domainProject][serviceName]*serviceCleanInfo
var svcCleanInfos = make(map[string]map[string]*serviceCleanInfo)

type serviceCleanInfo struct {
	ServiceName  string
	ServiceId    string
	WithInstance bool
	ShouldClear  bool
}

func getContextWith(domain string, project string) context.Context {
	return util.WithNoCache(util.SetDomainProject(context.Background(), domain, project))
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
	ctx := getContextWith(domain, project)
	svcResp, err := apt.ServiceAPI.Create(ctx, svc)
	Expect(err).To(BeNil())
	Expect(svcResp).NotTo(BeNil())
	Expect(svcResp.Response.GetCode()).To(Equal(pb.ResponseSuccess))
	info := &serviceCleanInfo{
		ServiceName:  name,
		ServiceId:    svcResp.ServiceId,
		WithInstance: withInstance,
		ShouldClear:  shouldClear,
	}
	domainProject := domain + path.SPLIT + project
	m, ok := svcCleanInfos[domainProject]
	if !ok {
		m = make(map[string]*serviceCleanInfo)
		svcCleanInfos[domainProject] = m
	}
	m[name] = info
}

func checkServiceCleared(domain string, project string) {
	domainProject := domain + path.SPLIT + project
	m := svcCleanInfos[domainProject]
	for _, v := range m {
		By(fmt.Sprintf("check cleared, service: %s, should be cleared: %t", v.ServiceName, v.ShouldClear))
		getSvcReq := &pb.GetServiceRequest{
			ServiceId: v.ServiceId,
		}
		ctx := getContextWith(domain, project)
		getSvcResp, err := apt.ServiceAPI.GetOne(ctx, getSvcReq)
		Expect(err).To(BeNil())
		Expect(getSvcResp).NotTo(BeNil())
		Expect(getSvcResp.Response.GetCode() == pb.ResponseSuccess).To(Equal(!v.ShouldClear))
	}
}
