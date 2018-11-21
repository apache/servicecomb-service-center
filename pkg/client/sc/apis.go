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
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/error"
	"github.com/apache/servicecomb-service-center/server/plugin/pkg/registry"
	"github.com/apache/servicecomb-service-center/version"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
)

const (
	apiVersionURL   = "/version"
	apiDumpURL      = "/v4/default/admin/dump"
	apiClustersURL  = "/v4/default/admin/clusters"
	apiHealthURL    = "/v4/default/registry/health"
	apiSchemasURL   = "/v4/%s/registry/microservices/%s/schemas"
	apiSchemaURL    = "/v4/%s/registry/microservices/%s/schemas/%s"
	apiInstancesURL = "/v4/%s/registry/microservices/%s/instances"
	apiInstanceURL  = "/v4/%s/registry/microservices/%s/instances/%s"

	QueryGlobal = "global"
)

func (c *SCClient) toError(body []byte) *scerr.Error {
	message := new(scerr.Error)
	err := json.Unmarshal(body, message)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, util.BytesToStringWithNoCopy(body))
	}
	return message
}

func (c *SCClient) parseQuery(ctx context.Context) (q string) {
	switch {
	case ctx.Value(QueryGlobal) == "1":
		q += "global=true"
	default:
		q += "global=false"
	}
	return
}

func (c *SCClient) GetScVersion(ctx context.Context) (*version.VersionSet, *scerr.Error) {
	resp, err := c.RestDoWithContext(ctx, http.MethodGet, apiVersionURL, c.CommonHeaders(ctx), nil)
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

	v := &version.VersionSet{}
	err = json.Unmarshal(body, v)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return v, nil
}

func (c *SCClient) GetScCache(ctx context.Context) (*model.Cache, *scerr.Error) {
	headers := c.CommonHeaders(ctx)
	// only default domain has admin permission
	headers.Set("X-Domain-Name", "default")
	resp, err := c.RestDoWithContext(ctx, http.MethodGet, apiDumpURL, headers, nil)
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

	dump := &model.DumpResponse{}
	err = json.Unmarshal(body, dump)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return dump.Cache, nil
}

func (c *SCClient) GetSchemasByServiceId(ctx context.Context, domainProject, serviceId string) ([]*pb.Schema, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiSchemasURL, project, serviceId)+"?withSchema=1&"+c.parseQuery(ctx),
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

	schemas := &pb.GetAllSchemaResponse{}
	err = json.Unmarshal(body, schemas)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return schemas.Schemas, nil
}

func (c *SCClient) GetSchemaBySchemaId(ctx context.Context, domainProject, serviceId, schemaId string) (*pb.Schema, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiSchemaURL, project, serviceId, schemaId)+"?"+c.parseQuery(ctx),
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

	schema := &pb.GetSchemaResponse{}
	err = json.Unmarshal(body, schema)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return &pb.Schema{
		SchemaId: schemaId,
		Schema:   schema.Schema,
		Summary:  schema.SchemaSummary,
	}, nil
}

func (c *SCClient) GetClusters(ctx context.Context) (registry.Clusters, *scerr.Error) {
	headers := c.CommonHeaders(ctx)
	// only default domain has admin permission
	headers.Set("X-Domain-Name", "default")
	resp, err := c.RestDoWithContext(ctx, http.MethodGet, apiClustersURL, headers, nil)
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

	clusters := &model.ClustersResponse{}
	err = json.Unmarshal(body, clusters)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return clusters.Clusters, nil
}

func (c *SCClient) HealthCheck(ctx context.Context) *scerr.Error {
	headers := c.CommonHeaders(ctx)
	// only default domain has admin permission
	headers.Set("X-Domain-Name", "default")
	resp, err := c.RestDoWithContext(ctx, http.MethodGet, apiHealthURL, headers, nil)
	if err != nil {
		return scerr.NewError(scerr.ErrUnavailableBackend, err.Error())
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
