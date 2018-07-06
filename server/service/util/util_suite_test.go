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
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/quota/buildin"
	_ "github.com/apache/incubator-servicecomb-service-center/server/plugin/infra/registry/buildin"
)

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
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
	_, err := serviceUtil.FindServiceIds(context.Background(),
		"latest", &proto.MicroServiceKey{})
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.FindServiceIds(context.Background(),
		"1.0.0", &proto.MicroServiceKey{})
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.FindServiceIds(context.Background(),
		"1.0+", &proto.MicroServiceKey{Alias: "test"})
	if err == nil {
		t.FailNow()
	}
}

func TestGetService(t *testing.T) {
	_, err := serviceUtil.GetService(context.Background(), "", "")
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServicesByDomainProject(context.Background(), "")
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetAllServiceUtil(context.Background())
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 0)
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 1)
	if err == nil {
		t.FailNow()
	}
}

func TestServiceExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.FailNow()
		}
	}()
	serviceUtil.ServiceExist(util.SetContext(context.Background(), serviceUtil.CTX_CACHEONLY, "1"), "", "")
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), serviceUtil.CTX_NOCACHE, "1")
	opts := serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.FailNow()
	}

	op := registry.OptionsToOp(opts...)
	if op.Mode != registry.MODE_NO_CACHE {
		t.FailNow()
	}

	ctx = context.WithValue(context.Background(), serviceUtil.CTX_CACHEONLY, "1")
	opts = serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.FailNow()
	}

	op = registry.OptionsToOp(opts...)
	if op.Mode != registry.MODE_CACHE {
		t.FailNow()
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
