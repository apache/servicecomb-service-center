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
package upgrade

import (
	"github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
	"strings"
)

func init() {
	AddPatch("0-0.1.1", ChangeIncompatibleKeysStore)
}

func ChangeIncompatibleKeysStore() error {
	// get all domain/project
	resp, err := registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
		Action:     registry.GET,
		Key:        []byte(core.GetDomainRootKey()),
		WithPrefix: true,
		KeyOnly:    true,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "get all domains failed")
		return err
	}

	domainProject := map[string]struct{}{}
	for _, kv := range resp.Kvs {
		key := registry.BytesToStringWithNoCopy(kv.Key)
		projResp, err := registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
			Action:     registry.GET,
			Key:        []byte(core.GetDomainProjectRootKey(key[strings.LastIndex(key, "/")+1:])),
			WithPrefix: true,
			KeyOnly:    true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "get all domain/projects failed")
			return err
		}
		for _, projKv := range projResp.Kvs {
			key := registry.BytesToStringWithNoCopy(projKv.Key)
			arr := strings.Split(key, "/")
			str := arr[1] + "/" + arr[2]
			if _, ok := domainProject[str]; !ok {
				domainProject[str] = struct{}{}
			}
		}
	}

	for domain := range domainProject {
		// tag
		resp, err := registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
			Action:     registry.GET,
			Key:        []byte(core.GetOldServiceTagRootKey(domain)),
			WithPrefix: true,
		})
		if err != nil {
			util.LOGGER.Errorf(err, "get all old tags failed")
			return err
		}

		for _, kv := range resp.Kvs {
			key := registry.BytesToStringWithNoCopy(kv.Key)
			serviceId := key[strings.LastIndex(key, "/")+1:]
			_, err := registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
				Action: registry.PUT,
				Key:    []byte(core.GenerateServiceTagKey(domain, serviceId)),
				Value:  kv.Value,
			})
			if err != nil {
				util.LOGGER.Errorf(err, "put new tags failed")
				return err
			}
		}
	}

	util.LOGGER.Infof("patch ChangeIncompatibleKeysStore(0.1.1) install [OK]")
	return nil
}
