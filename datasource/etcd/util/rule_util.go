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

package util

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

func AllowAcrossDimension(ctx context.Context, providerService *discovery.MicroService, consumerService *discovery.MicroService) error {
	if providerService.AppId != consumerService.AppId {
		if len(providerService.Properties) == 0 {
			return fmt.Errorf("not allow across app access")
		}

		if allowCrossApp, ok := providerService.Properties[discovery.PropAllowCrossApp]; !ok || strings.ToLower(allowCrossApp) != "true" {
			return fmt.Errorf("not allow across app access")
		}
	}

	if !datasource.IsGlobal(discovery.MicroServiceToKey(util.ParseTargetDomainProject(ctx), providerService)) &&
		providerService.Environment != consumerService.Environment {
		return fmt.Errorf("not allow across environment access")
	}

	return nil
}

func Accessible(ctx context.Context, consumerID string, providerID string) *errsvc.Error {
	if len(consumerID) == 0 {
		return nil
	}

	domainProject := util.ParseDomainProject(ctx)
	targetDomainProject := util.ParseTargetDomainProject(ctx)

	consumerService, err := GetService(ctx, domainProject, consumerID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return discovery.NewError(discovery.ErrServiceNotExists, "consumer serviceID is invalid")
		}
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("An error occurred in query consumer(%s)", err.Error()))
	}

	// 跨应用权限
	providerService, err := GetService(ctx, targetDomainProject, providerID)
	if err != nil {
		if errors.Is(err, datasource.ErrNoData) {
			return discovery.NewError(discovery.ErrServiceNotExists, "provider serviceID is invalid")
		}
		return discovery.NewError(discovery.ErrInternal, fmt.Sprintf("An error occurred in query provider(%s)", err.Error()))
	}

	err = AllowAcrossDimension(ctx, providerService, consumerService)
	if err != nil {
		return discovery.NewError(discovery.ErrPermissionDeny, err.Error())
	}
	return nil
}
