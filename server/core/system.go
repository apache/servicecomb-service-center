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
package core

import (
	"encoding/json"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
)

var systemConfig *SystemConfig

type SystemConfig struct {
	Version string `json:"version"`
}

func LoadSystemConfig() error {
	resp, err := registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(GetSystemKey()),
	})
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		systemConfig = &SystemConfig{
			Version: "0",
		}
		return nil
	}
	systemConfig = &SystemConfig{
		Version: "0",
	}
	err = json.Unmarshal(resp.Kvs[0].Value, systemConfig)
	if err != nil {
		util.LOGGER.Errorf(err, "load system config failed, maybe incompatible")
		return nil
	}
	return nil
}

func UpgradeSystemConfig() error {
	bytes, err := json.Marshal(GetSystemConfig())
	if err != nil {
		return err
	}
	_, err = registry.GetRegisterCenter().Do(context.Background(), &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(GetSystemKey()),
		Value:  bytes,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetSystemConfig() *SystemConfig {
	return systemConfig
}
