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

package clusters

import (
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/discovery"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"golang.org/x/net/context"
)

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "clusters", New})
}

var discoveries = make(map[string]discovery.Discovery)

type Clusters []discovery.Discovery

func (c Clusters) MicroServices(ctx context.Context, opts ...registry.PluginOpOption) (arr []*pb.MicroService, err error) {
	for _, d := range c {
		ms, err := d.MicroServices(ctx, opts...)
		if err != nil {
			return nil, err
		}
		arr = append(arr, ms...)
	}
	return
}

func (c Clusters) Instances(ctx context.Context, opts ...registry.PluginOpOption) (arr []*pb.MicroServiceInstance, err error) {
	for _, d := range c {
		ms, err := d.Instances(ctx, opts...)
		if err != nil {
			return nil, err
		}
		arr = append(arr, ms...)
	}
	return
}

// New must be called after all discovery registered
func New() mgr.PluginInstance {
	var i Clusters
	for _, d := range discoveries {
		i = append(i, d)
	}
	return i
}

func RegisterDiscovery(name string, discovery discovery.Discovery) {
	discoveries[name] = discovery
}
