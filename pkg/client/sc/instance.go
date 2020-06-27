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

package sc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

const (
	apiDiscoveryInstancesURL = "/v4/%s/registry/instances"
	apiHeartbeatSetURL       = "/v4/%s/registry/heartbeats"
	apiInstancesURL          = "/v4/%s/registry/microservices/%s/instances"
	apiInstanceURL           = "/v4/%s/registry/microservices/%s/instances/%s"
	apiInstanceHeartbeatURL  = "/v4/%s/registry/microservices/%s/instances/%s/heartbeat"
)

func (c *SCClient) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *pb.MicroServiceInstance) (string, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.RegisterInstanceRequest{Instance: instance})
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiInstancesURL, project, serviceId),
		headers, reqBody)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.toError(body)
	}

	instancesResp := &pb.RegisterInstanceResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return instancesResp.InstanceId, nil
}

func (c *SCClient) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiInstanceURL, project, serviceId, instanceId),
		headers, nil)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *SCClient) DiscoveryInstances(ctx context.Context, domainProject, consumerId, providerAppId, providerServiceName, providerVersionRule string) ([]*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerId)

	query := url.Values{}
	query.Set("appId", providerAppId)
	query.Set("serviceName", providerServiceName)
	query.Set("version", providerVersionRule)

	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiDiscoveryInstancesURL, project)+"?"+c.parseQuery(ctx)+"&"+query.Encode(),
		headers, nil)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &pb.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *SCClient) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiInstanceHeartbeatURL, project, serviceId, instanceId),
		headers, nil)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *SCClient) HeartbeatSet(ctx context.Context, domainProject string, instances ...*pb.HeartbeatSetElement) ([]*pb.InstanceHbRst, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.HeartbeatSetRequest{Instances: instances})
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiHeartbeatSetURL, project),
		headers, reqBody)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &pb.HeartbeatSetResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *SCClient) GetInstancesByServiceId(ctx context.Context, domainProject, providerId, consumerId string) ([]*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerId)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstancesURL, project, providerId)+"?"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &pb.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *SCClient) GetInstanceByInstanceId(ctx context.Context, domainProject, providerId, instanceId, consumerId string) (*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerId)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstanceURL, project, providerId, instanceId)+"?"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instanceResp := &pb.GetOneInstanceResponse{}
	err = json.Unmarshal(body, instanceResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instanceResp.Instance, nil
}
