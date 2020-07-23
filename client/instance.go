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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"io/ioutil"
	"net/http"
	"net/url"

	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

const (
	apiDiscoveryInstancesURL = "/v4/%s/registry/instances"
	apiHeartbeatSetURL       = "/v4/%s/registry/heartbeats"
	apiInstancesURL          = "/v4/%s/registry/microservices/%s/instances"
	apiInstanceURL           = "/v4/%s/registry/microservices/%s/instances/%s"
	apiInstanceHeartbeatURL  = "/v4/%s/registry/microservices/%s/instances/%s/heartbeat"
)

func (c *Client) RegisterInstance(ctx context.Context, domain, project, serviceID string, instance *registry.MicroServiceInstance) (string, *scerr.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&registry.RegisterInstanceRequest{Instance: instance})
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

	instancesResp := &registry.RegisterInstanceResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return instancesResp.InstanceId, nil
}

func (c *Client) UnregisterInstance(ctx context.Context, domain, project, serviceID, instanceID string) *scerr.Error {
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

func (c *Client) DiscoveryInstances(ctx context.Context, domain, project, consumerID, providerAppID, providerServiceName, providerVersionRule string) ([]*registry.MicroServiceInstance, *scerr.Error) {
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

	instancesResp := &registry.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) Heartbeat(ctx context.Context, domain, project, serviceID, instanceID string) *scerr.Error {
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

func (c *Client) HeartbeatSet(ctx context.Context, domain, project string, instances ...*registry.HeartbeatSetElement) ([]*registry.InstanceHbRst, *scerr.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&registry.HeartbeatSetRequest{Instances: instances})
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

	instancesResp := &registry.HeartbeatSetResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) GetInstancesByServiceID(ctx context.Context, domain, project, providerID, consumerID string) ([]*registry.MicroServiceInstance, *scerr.Error) {
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

	instancesResp := &registry.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) GetInstanceByInstanceID(ctx context.Context, domain, project, providerID, instanceID, consumerID string) (*registry.MicroServiceInstance, *scerr.Error) {
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

	instanceResp := &registry.GetOneInstanceResponse{}
	err = json.Unmarshal(body, instanceResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instanceResp.Instance, nil
}
