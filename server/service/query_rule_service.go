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
package service

import (
	"fmt"
	ms "github.com/ServiceComb/service-center/server/service/microservice"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"github.com/ServiceComb/service-center/util"
	errorsEx "github.com/ServiceComb/service-center/util/errors"
	"golang.org/x/net/context"
)

func Accessible(ctx context.Context, tenant string, consumerId string, providerId string) error {
	consumerService, err := ms.GetServiceByServiceId(ctx, tenant, consumerId)
	if err != nil {
		util.LOGGER.Errorf(err,
			"consumer %s can't access provider %s for internal error", consumerId, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if consumerService == nil {
		util.LOGGER.Warnf(nil,
			"consumer %s can't access provider %s for invalid consumer", consumerId, providerId)
		return fmt.Errorf("consumer invalid")
	}

	consumerFlag := fmt.Sprintf("%s/%s/%s", consumerService.AppId, consumerService.ServiceName, consumerService.Version)

	// 跨应用权限
	providerService, err := ms.GetServiceByServiceId(ctx, tenant, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerId)
		return errorsEx.InternalError(err.Error())
	}
	if providerService == nil {
		util.LOGGER.Warnf(nil, "consumer %s can't access provider %s for invalid provider",
			consumerFlag, providerId)
		return fmt.Errorf("provider invalid")
	}

	providerFlag := fmt.Sprintf("%s/%s/%s", providerService.AppId, providerService.ServiceName, providerService.Version)

	err = serviceUtil.AllowAcrossApp(providerService, consumerService)
	if err != nil {
		util.LOGGER.Warnf(nil,
			"consumer %s can't access provider %s which property 'allowCrossApp' is not true or does not exist",
			consumerFlag, providerFlag)
		return err
	}

	// 黑白名单
	rules, err := serviceUtil.GetRulesUtil(ctx, tenant, providerId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	if len(rules) == 0 {
		return nil
	}

	validateTags, err := serviceUtil.GetTagsUtils(ctx, tenant, consumerService.ServiceId)
	if err != nil {
		util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
			consumerFlag, providerFlag)
		return errorsEx.InternalError(err.Error())
	}

	err = serviceUtil.MatchRules(rules, consumerService, validateTags)
	if err != nil {
		switch err.(type) {
		case errorsEx.InternalError:
			util.LOGGER.Errorf(err, "consumer %s can't access provider %s for internal error",
				consumerFlag, providerFlag)
		default:
			util.LOGGER.Warnf(err, "consumer %s can't access provider %s", consumerFlag, providerFlag)
		}
		return err
	}

	return nil
}
