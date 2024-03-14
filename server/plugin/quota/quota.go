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

package quota

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
)

const QUOTA plugin.Kind = "quota"

type Manager interface {
	RemandQuotas(ctx context.Context, t ResourceType)
	GetQuota(ctx context.Context, t ResourceType) int64
	Usage(ctx context.Context, req *Request) (int64, error)
}

func GetQuota(ctx context.Context, resourceType ResourceType) int64 {
	return plugin.Plugins().Instance(QUOTA).(Manager).GetQuota(ctx, resourceType)
}

func Remand(ctx context.Context, resourceType ResourceType) {
	plugin.Plugins().Instance(QUOTA).(Manager).RemandQuotas(ctx, resourceType)
}

func Usage(ctx context.Context, req *Request) (int64, error) {
	return plugin.Plugins().Instance(QUOTA).(Manager).Usage(ctx, req)
}

// Apply 申请配额sourceType serviceinstance servicetype
func Apply(ctx context.Context, res *Request) error {
	if res == nil {
		err := errors.New("invalid parameters")
		log.Error("quota check failed", err)
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	resourceType := res.QuotaType
	limitQuota := GetQuota(ctx, resourceType)
	curNum, err := Usage(ctx, res)
	if err != nil {
		log.Error(fmt.Sprintf("%s quota check failed", resourceType), err)
		return err
	}
	if curNum+res.QuotaSize > limitQuota {
		mes := fmt.Sprintf("no quota to create %s, max num is %d, curNum is %d, apply num is %d",
			resourceType, limitQuota, curNum, res.QuotaSize)
		log.Error(mes, nil)
		return pb.NewError(pb.ErrNotEnoughQuota, mes)
	}
	return nil
}
