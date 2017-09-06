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
package buildin

import (
	"fmt"
	constKey "github.com/ServiceComb/service-center/server/common"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	rl "github.com/ServiceComb/service-center/server/infra/resourceslimit"
	"github.com/ServiceComb/service-center/util"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/net/context"
)

type BuildInLimit struct {
}

func init() {
	rl.ResourcesLimitPlugins["buildin"] = New
}

func (buildin *BuildInLimit) CanApplyResource(resourceType rl.ResourceType, tenant string, serviceId string, applyCount int32) (bool, error) {
	var indexer *store.Indexer
	switch resourceType {
	case rl.RULE:
		indexer = store.Store().Rule()
		key := core.GenerateServiceRuleKey(tenant, serviceId, "")
		resp, err := indexer.Search(context.TODO(), &registry.PluginOp{
			Action:     registry.GET,
			Key:        util.StringToBytesWithNoCopy(key),
			CountOnly:  true,
			WithPrefix: true,
			Mode:       registry.MODE_NO_CACHE,
		})
		if err != nil {
			util.Logger().Errorf(err, "get service rule failed for check resources size, %s", serviceId)
			return false, err
		}
		num := resp.Count + int64(applyCount)
		if num > constKey.RULE_NUM_MAX_FOR_ONESERVICE {
			util.Logger().Errorf(nil, "fail to add rule for one service max rule num is %d, %s", constKey.RULE_NUM_MAX_FOR_ONESERVICE, serviceId)
			return false, fmt.Errorf("fail to add rule for one service max rule num is %d", constKey.RULE_NUM_MAX_FOR_ONESERVICE)
		}
	case rl.SCHEMA:
		key := core.GenerateServiceSchemaKey(tenant, serviceId, "")
		indexer = store.Store().Schema()
		resp, err := indexer.Search(context.TODO(), &registry.PluginOp{
			Action:     registry.GET,
			Key:        util.StringToBytesWithNoCopy(key),
			WithPrefix: true,
			CountOnly:  true,
			Mode:       registry.MODE_NO_CACHE,
		})
		if err != nil {
			return false, err
		}
		num := resp.Count + int64(applyCount)
		if num > constKey.SCHEMA_NUM_MAX_FOR_ONESERVICE {
			util.Logger().Errorf(nil, "fail to add schema for one service max rule num is %d, %s", constKey.SCHEMA_NUM_MAX_FOR_ONESERVICE, serviceId)
			return false, fmt.Errorf("fail to add schema for one service max rule num is %d", constKey.SCHEMA_NUM_MAX_FOR_ONESERVICE)
		}
	default:
		util.Logger().Errorf(nil, "wrong resource type, not in rule, tag, shema")
		return false, errors.New("wrong resource type, not in rule, tag, shema")

	}
	return true, nil
}

func New() rl.ResourcesManager {
	return &BuildInLimit{}
}
