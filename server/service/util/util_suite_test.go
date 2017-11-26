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
package util_test

import _ "github.com/ServiceComb/service-center/server/plugin/infra/registry/buildin"

import (
	"fmt"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/infra/registry"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/coreos/etcd/mvcc/mvccpb"
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
	_, err := serviceUtil.FindServiceIds(util.SetContext(context.Background(), "cacheOnly", "1"),
		"latest", &proto.MicroServiceKey{})
	if err != nil {
		t.FailNow()
	}

	_, err = serviceUtil.FindServiceIds(util.SetContext(context.Background(), "cacheOnly", "1"),
		"1.0.0", &proto.MicroServiceKey{})
	if err != nil {
		t.FailNow()
	}

	_, err = serviceUtil.FindServiceIds(util.SetContext(context.Background(), "cacheOnly", "1"),
		"1.0+", &proto.MicroServiceKey{Alias: "test"})
	if err != nil {
		t.FailNow()
	}
}

func TestGetService(t *testing.T) {
	_, err := serviceUtil.GetService(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
	if err != nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetService(context.Background(), "", "")
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServicesRawData(util.SetContext(context.Background(), "cacheOnly", "1"), "")
	if err != nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServicesRawData(context.Background(), "")
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServicesByDomain(util.SetContext(context.Background(), "cacheOnly", "1"), "")
	if err != nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetServicesByDomain(context.Background(), "")
	if err == nil {
		t.FailNow()
	}

	_, err = serviceUtil.GetAllServiceUtil(util.SetContext(context.Background(), "cacheOnly", "1"))
	if err != nil {
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

func TestMsCache(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.FailNow()
		}
	}()
	_, err := serviceUtil.GetServiceInCache(context.Background(), "", "")
	if err == nil {
		t.FailNow()
	}
	ms := serviceUtil.MsCache()
	if ms == nil {
		t.FailNow()
	}
	ms.Set("", &proto.MicroService{}, 0)
	_, err = serviceUtil.GetServiceInCache(context.Background(), "", "")
	if err != nil {
		t.FailNow()
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "noCache", "1")
	opts := serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.FailNow()
	}

	op := registry.OptionsToOp(opts...)
	if op.Mode != registry.MODE_NO_CACHE {
		t.FailNow()
	}

	ctx = context.WithValue(context.Background(), "cacheOnly", "1")
	opts = serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.FailNow()
	}

	op = registry.OptionsToOp(opts...)
	if op.Mode != registry.MODE_CACHE {
		t.FailNow()
	}
}

func TestServiceExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.FailNow()
		}
	}()
	serviceUtil.ServiceExist(util.SetContext(context.Background(), "cacheOnly", "1"), "", "")
}

const VERSIONRULE_BASE = 5000

func BenchmarkVersionRule_Latest_GetServicesIds(b *testing.B) {
	var kvs = make([]*mvccpb.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &mvccpb.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceUtil.VersionRule(serviceUtil.Latest).Match(kvs)
	}
	b.ReportAllocs()
	// 5000	   7105020 ns/op	 2180198 B/op	   39068 allocs/op
}

func BenchmarkVersionRule_Range_GetServicesIds(b *testing.B) {
	var kvs = make([]*mvccpb.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &mvccpb.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceUtil.VersionRule(serviceUtil.Range).Match(kvs, fmt.Sprintf("1.%d", i), fmt.Sprintf("1.%d", i+VERSIONRULE_BASE/10))
	}
	b.ReportAllocs()
	// 5000	   7244029 ns/op	 2287389 B/op	   39584 allocs/op
}

func BenchmarkVersionRule_AtLess_GetServicesIds(b *testing.B) {
	var kvs = make([]*mvccpb.KeyValue, VERSIONRULE_BASE)
	for i := 1; i <= VERSIONRULE_BASE; i++ {
		kvs[i-1] = &mvccpb.KeyValue{
			Key:   []byte(fmt.Sprintf("/service/ver/1.%d", i)),
			Value: []byte(fmt.Sprintf("%d", i)),
		}
	}
	b.N = VERSIONRULE_BASE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceUtil.VersionRule(serviceUtil.AtLess).Match(kvs, fmt.Sprintf("1.%d", i))
	}
	b.ReportAllocs()
	// 5000	  11221098 ns/op	 3174720 B/op	   58064 allocs/op
}
