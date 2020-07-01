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

func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceID string, instance *pb.MicroServiceInstance) (string, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.RegisterInstanceRequest{Instance: instance})
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiInstancesURL, project, serviceID),
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

func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceID, instanceID string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiInstanceURL, project, serviceID, instanceID),
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

func (c *Client) DiscoveryInstances(ctx context.Context, domainProject, consumerID, providerAppID, providerServiceName, providerVersionRule string) ([]*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerID)

	query := url.Values{}
	query.Set("appId", providerAppID)
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

func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceID, instanceID string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiInstanceHeartbeatURL, project, serviceID, instanceID),
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

func (c *Client) HeartbeatSet(ctx context.Context, domainProject string, instances ...*pb.HeartbeatSetElement) ([]*pb.InstanceHbRst, *scerr.Error) {
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

func (c *Client) GetInstancesByServiceID(ctx context.Context, domainProject, providerID, consumerID string) ([]*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerID)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstancesURL, project, providerID)+"?"+c.parseQuery(ctx),
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

func (c *Client) GetInstanceByInstanceID(ctx context.Context, domainProject, providerID, instanceID, consumerID string) (*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerID)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstanceURL, project, providerID, instanceID)+"?"+c.parseQuery(ctx),
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
