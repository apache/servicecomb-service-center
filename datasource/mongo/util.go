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

package mongo

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/util"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
)

type InstanceSlice []*pb.MicroServiceInstance

func (s InstanceSlice) Len() int {
	return len(s)
}

func (s InstanceSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s InstanceSlice) Less(i, j int) bool {
	return s[i].InstanceId < s[j].InstanceId
}

func statistics(ctx context.Context, withShared bool) (*pb.Statistics, error) {
	result := &pb.Statistics{
		Services:  &pb.StService{},
		Instances: &pb.StInstance{},
		Apps:      &pb.StApp{},
	}

	services, err := GetAllMicroServicesByDomainProject(ctx)
	if err != nil {
		return nil, err
	}

	var svcIDs []string
	var svcKeys []*pb.MicroServiceKey
	for _, svc := range services {
		svcIDs = append(svcIDs, svc.ServiceId)
		svcKeys = append(svcKeys, datasource.TransServiceToKey(util.ParseDomainProject(ctx), svc))
	}
	svcIDToNonVerKey := datasource.SetStaticServices(result, svcKeys, svcIDs, withShared)
	respGetInstanceCountByDomain := make(chan datasource.GetInstanceCountByDomainResponse, 1)
	gopool.Go(func(_ context.Context) {
		getInstanceCountByDomain(ctx, svcIDToNonVerKey, respGetInstanceCountByDomain)
	})

	instances, err := GetInstances(ctx)
	if err != nil {
		return nil, err
	}
	var instIDs []string
	for _, inst := range instances {
		instIDs = append(instIDs, inst.Instance.ServiceId)
	}
	datasource.SetStaticInstances(result, svcIDToNonVerKey, instIDs)
	data := <-respGetInstanceCountByDomain
	close(respGetInstanceCountByDomain)
	if data.Err != nil {
		return nil, data.Err
	}
	result.Instances.CountByDomain = data.CountByDomain
	return result, nil
}

func getInstanceCountByDomain(ctx context.Context, svcIDToNonVerKey map[string]string, resp chan datasource.GetInstanceCountByDomainResponse) {
	ret := datasource.GetInstanceCountByDomainResponse{}
	for sid := range svcIDToNonVerKey {
		num, err := CountInstance(ctx, sid)
		if err != nil {
			ret.Err = err
			return
		}
		ret.CountByDomain = ret.CountByDomain + num
	}
	resp <- ret
}
