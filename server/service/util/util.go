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
package util

import "github.com/ServiceComb/service-center/server/core/registry"

type QueryOp func() []registry.PluginOpOption

func WithNoCache(no bool) QueryOp {
	if !no {
		return func() []registry.PluginOpOption { return nil }
	}
	return func() []registry.PluginOpOption {
		return []registry.PluginOpOption{registry.WithNoCache()}
	}
}

func QueryOptions(qopts ...QueryOp) (opts []registry.PluginOpOption) {
	if len(qopts) == 0 {
		return
	}
	opts = []registry.PluginOpOption{}
	for _, qopt := range qopts {
		opts = append(opts, qopt()...)
	}
	return
}
