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

package buildin

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"testing"
)

func TestCommonQuotaCheck(t *testing.T) {
	// case: invalid input
	rst := CommonQuotaCheck(context.Background(), nil, func() int64 {
		return 1
	}, func(ctx context.Context, resource *quota.ApplyQuotaResource) (int64, error) {
		return 0, nil
	})
	if rst.Err == nil || !rst.Err.InternalError() {
		t.Fatalf("TestCommonQuotaCheck failed")
	}
	rst = CommonQuotaCheck(context.Background(), &quota.ApplyQuotaResource{}, nil, func(ctx context.Context, resource *quota.ApplyQuotaResource) (int64, error) {
		return 0, nil
	})
	if rst.Err == nil || !rst.Err.InternalError() {
		t.Fatalf("TestCommonQuotaCheck failed")
	}
	rst = CommonQuotaCheck(context.Background(), &quota.ApplyQuotaResource{}, func() int64 {
		return 1
	}, nil)
	if rst.Err == nil || !rst.Err.InternalError() {
		t.Fatalf("TestCommonQuotaCheck failed")
	}

	// case: error
	rst = CommonQuotaCheck(context.Background(), &quota.ApplyQuotaResource{
		QuotaType: quota.MicroServiceQuotaType,
		QuotaSize: 1,
	}, func() int64 {
		return 1
	}, func(_ context.Context, _ *quota.ApplyQuotaResource) (int64, error) {
		return 0, errors.New("error")
	})
	if rst.Err == nil || !rst.Err.InternalError() {
		t.Fatalf("TestCommonQuotaCheck failed %v", rst.Err)
	}

	// case: normal
	rst = CommonQuotaCheck(context.Background(), &quota.ApplyQuotaResource{
		QuotaType: quota.MicroServiceQuotaType,
		QuotaSize: 1,
	}, func() int64 {
		return 1
	}, func(_ context.Context, _ *quota.ApplyQuotaResource) (int64, error) {
		return 0, nil
	})
	if rst.Err != nil {
		t.Fatalf("TestCommonQuotaCheck failed %v", rst.Err)
	}

	rst = CommonQuotaCheck(context.Background(), &quota.ApplyQuotaResource{
		QuotaType: quota.MicroServiceQuotaType,
		QuotaSize: 1,
	}, func() int64 {
		return 1
	}, func(_ context.Context, _ *quota.ApplyQuotaResource) (int64, error) {
		return 1, nil
	})
	if rst.Err == nil || rst.Err.InternalError() {
		t.Fatalf("TestCommonQuotaCheck failed %v", rst.Err)
	}
}
