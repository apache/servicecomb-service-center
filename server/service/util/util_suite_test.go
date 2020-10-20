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
package util_test

import (
	_ "github.com/apache/servicecomb-service-center/datasource/etcd/client/buildin"
	_ "github.com/apache/servicecomb-service-center/datasource/etcd/sd/etcd"
	_ "github.com/apache/servicecomb-service-center/server/plugin/quota/buildin"
)

import (
	"context"
	proto "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
}

func TestMicroservice(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("model.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "model Suite", []Reporter{junitReporter})
}

func TestFindServiceIds(t *testing.T) {
	_, _, err := serviceUtil.FindServiceIds(context.Background(),
		"latest", &proto.MicroServiceKey{})
	if err != nil {
		t.Fatalf("TestFindServiceIds failed")
	}

	_, _, err = serviceUtil.FindServiceIds(context.Background(),
		"1.0.0", &proto.MicroServiceKey{})
	if err != nil {
		t.Fatalf("TestFindServiceIds failed")
	}

	_, _, err = serviceUtil.FindServiceIds(context.Background(),
		"1.0+", &proto.MicroServiceKey{Alias: "test"})
	if err != nil {
		t.Fatalf("TestFindServiceIds failed")
	}
}

func TestGetService(t *testing.T) {
	_, err := serviceUtil.GetService(context.Background(), "", "")
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetServicesByDomainProject(context.Background(), "")
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetAllServiceUtil(context.Background())
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 0)
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 1)
	if err != nil {
		t.Fatalf("TestGetService failed")
	}
}

func TestServiceExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestServiceExist failed")
		}
	}()
	serviceUtil.ServiceExist(util.SetContext(context.Background(), util.CtxCacheOnly, "1"), "", "")
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), util.CtxNocache, "1")
	opts := serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op := registry.OptionsToOp(opts...)
	if op.Mode != registry.ModeNoCache {
		t.Fatalf("TestFromContext failed")
	}

	ctx = context.WithValue(context.Background(), util.CtxCacheOnly, "1")
	opts = serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op = registry.OptionsToOp(opts...)
	if op.Mode != registry.ModeCache {
		t.Fatalf("TestFromContext failed")
	}
}

func TestRemandQuota(t *testing.T) {
	serviceUtil.RemandServiceQuota(context.Background())
	serviceUtil.RemandInstanceQuota(context.Background())
}

func TestSetDefault(t *testing.T) {
	service := &proto.MicroService{}
	serviceUtil.SetServiceDefaultValue(service)
	if len(service.Level) == 0 ||
		len(service.Status) == 0 {
		t.Fatalf(`TestSetDefault failed`)
	}
}
