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

// initialize
import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	_ "github.com/apache/servicecomb-service-center/test"

	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/util"

	proto "github.com/go-chassis/cari/discovery"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/gomega"
)

func init() {
}

func TestMicroservice(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("util.junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "util Suite", []Reporter{junitReporter})
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
	var err error
	t.Run("when there is no such a service in db", func(t *testing.T) {
		_, err = serviceUtil.GetService(context.Background(), "", "")
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("TestGetService failed")
		}
		assert.Equal(t, datasource.ErrNoData, err)
	})

	_, err = serviceUtil.GetServicesByDomainProject(context.Background(), "")
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	_, err = serviceUtil.GetAllServiceUtil(context.Background())
	if err != nil {
		t.Fatalf("TestGetService failed")
	}

	t.Run("when there is no such a service in db", func(t *testing.T) {
		_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 0)
		if err != nil && !errors.Is(err, datasource.ErrNoData) {
			t.Fatalf("TestGetService failed")
		}
		assert.Equal(t, datasource.ErrNoData, err)
	})

	//_, err = serviceUtil.GetServiceWithRev(context.Background(), "", "", 1)
	//if err != nil {
	//	t.Fatalf("TestGetService failed")
	//}
}

func TestServiceExist(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TestServiceExist failed")
		}
	}()
	serviceUtil.ServiceExist(util.WithCacheOnly(context.Background()), "", "")
}

func TestFromContext(t *testing.T) {
	ctx := util.WithNoCache(context.Background())
	opts := serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op := client.OptionsToOp(opts...)
	if op.Mode != client.ModeNoCache {
		t.Fatalf("TestFromContext failed")
	}

	ctx = util.WithCacheOnly(context.Background())
	opts = serviceUtil.FromContext(ctx)
	if len(opts) == 0 {
		t.Fatalf("TestFromContext failed")
	}

	op = client.OptionsToOp(opts...)
	if op.Mode != client.ModeCache {
		t.Fatalf("TestFromContext failed")
	}
}

func TestRemandQuota(t *testing.T) {
	serviceUtil.RemandServiceQuota(context.Background())
	serviceUtil.RemandInstanceQuota(context.Background())
}
