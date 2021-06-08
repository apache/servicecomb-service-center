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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	apiDiscoveryInstancesURL = "/v4/%s/registry/instances"
	apiHeartbeatSetURL       = "/v4/%s/registry/heartbeats"
	apiInstancesURL          = "/v4/%s/registry/microservices/%s/instances"
	apiInstanceURL           = "/v4/%s/registry/microservices/%s/instances/%s"
	apiInstanceHeartbeatURL  = "/v4/%s/registry/microservices/%s/instances/%s/heartbeat"
)

func (c *Client) RegisterInstance(ctx context.Context, domain, project, serviceID string, instance *discovery.MicroServiceInstance) (string, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&discovery.RegisterInstanceRequest{Instance: instance})
	if err != nil {
		return "", discovery.NewError(discovery.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiInstancesURL, project, serviceID),
		headers, reqBody)
	if err != nil {
		return "", discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.toError(body)
	}

	instancesResp := &discovery.RegisterInstanceResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return "", discovery.NewError(discovery.ErrInternal, err.Error())
	}
	return instancesResp.InstanceId, nil
}

func (c *Client) UnregisterInstance(ctx context.Context, domain, project, serviceID, instanceID string) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiInstanceURL, project, serviceID, instanceID),
		headers, nil)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *Client) DiscoveryInstances(ctx context.Context, domain, project, consumerID, providerAppID, providerServiceName, providerVersionRule string) ([]*discovery.MicroServiceInstance, *errsvc.Error) {
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
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &discovery.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) Heartbeat(ctx context.Context, domain, project, serviceID, instanceID string) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiInstanceHeartbeatURL, project, serviceID, instanceID),
		headers, nil)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *Client) HeartbeatSet(ctx context.Context, domain, project string, instances ...*discovery.HeartbeatSetElement) ([]*discovery.InstanceHbRst, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&discovery.HeartbeatSetRequest{Instances: instances})
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiHeartbeatSetURL, project),
		headers, reqBody)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &discovery.HeartbeatSetResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) GetInstancesByServiceID(ctx context.Context, domain, project, providerID, consumerID string) ([]*discovery.MicroServiceInstance, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerID)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstancesURL, project, providerID)+"?"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &discovery.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *Client) GetInstanceByInstanceID(ctx context.Context, domain, project, providerID, instanceID, consumerID string) (*discovery.MicroServiceInstance, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerID)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiInstanceURL, project, providerID, instanceID)+"?"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instanceResp := &discovery.GetOneInstanceResponse{}
	err = json.Unmarshal(body, instanceResp)
	if err != nil {
		return nil, discovery.NewError(discovery.ErrInternal, err.Error())
	}

	return instanceResp.Instance, nil
}
