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
	scerr "github.com/apache/servicecomb-service-center/server/error"
)

const (
	apiExistenceURL     = "/v4/%s/registry/existence"
	apiMicroServicesURL = "/v4/%s/registry/microservices"
	apiMicroServiceURL  = "/v4/%s/registry/microservices/%s"
)

func (c *SCClient) CreateService(ctx context.Context, domainProject string, service *pb.MicroService) (string, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.CreateServiceRequest{Service: service})
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiMicroServicesURL, project),
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

	serviceResp := &pb.CreateServiceResponse{}
	err = json.Unmarshal(body, serviceResp)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return serviceResp.ServiceId, nil
}

func (c *SCClient) DeleteService(ctx context.Context, domainProject, serviceId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiMicroServiceURL, project, serviceId),
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

func (c *SCClient) ServiceExistence(ctx context.Context, domainProject string, appId, serviceName, versionRule, env string) (string, *scerr.Error) {
	query := url.Values{}
	query.Set("type", "microservice")
	query.Set("env", env)
	query.Set("appId", appId)
	query.Set("serviceName", serviceName)
	query.Set("version", versionRule)

	resp, err := c.existence(ctx, domainProject, query)
	if err != nil {
		return "", err
	}

	return resp.ServiceId, nil
}

func (c *SCClient) existence(ctx context.Context, domainProject string, query url.Values) (*pb.GetExistenceResponse, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiExistenceURL, project)+"?"+c.parseQuery(ctx)+"&"+query.Encode(),
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

	existenceResp := &pb.GetExistenceResponse{}
	err = json.Unmarshal(body, existenceResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return existenceResp, nil
}
