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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	apiExistenceURL     = "/v4/%s/registry/existence"
	apiMicroServicesURL = "/v4/%s/registry/microservices"
	apiMicroServiceURL  = "/v4/%s/registry/microservices/%s"
)

func (c *Client) CreateService(ctx context.Context, domain, project string, service *pb.MicroService) (string, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.CreateServiceRequest{Service: service})
	if err != nil {
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiMicroServicesURL, project),
		headers, reqBody)
	if err != nil {
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.toError(body)
	}

	serviceResp := &pb.CreateServiceResponse{}
	err = json.Unmarshal(body, serviceResp)
	if err != nil {
		return "", pb.NewError(pb.ErrInternal, err.Error())
	}
	return serviceResp.ServiceId, nil
}

func (c *Client) DeleteService(ctx context.Context, domain, project, serviceID string) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiMicroServiceURL, project, serviceID),
		headers, nil)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *Client) ServiceExistence(ctx context.Context, domain, project string, appID, serviceName, versionRule, env string) (string, *errsvc.Error) {
	query := url.Values{}
	query.Set("type", "microservice")
	query.Set("env", env)
	query.Set("appId", appID)
	query.Set("serviceName", serviceName)
	query.Set("version", versionRule)

	resp, err := c.existence(ctx, domain, project, query)
	if err != nil {
		return "", err
	}

	return resp.ServiceId, nil
}

func (c *Client) existence(ctx context.Context, domain, project string, query url.Values) (*pb.GetExistenceResponse, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiExistenceURL, project)+"?"+c.parseQuery(ctx)+"&"+query.Encode(),
		headers, nil)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	existenceResp := &pb.GetExistenceResponse{}
	err = json.Unmarshal(body, existenceResp)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return existenceResp, nil
}
