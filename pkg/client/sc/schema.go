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

package sc

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
)

const (
	apiSchemasURL = "/v4/%s/registry/microservices/%s/schemas"
	apiSchemaURL  = "/v4/%s/registry/microservices/%s/schemas/%s"
)

func (c *SCClient) CreateSchemas(ctx context.Context, domainProject, serviceId string, schemas []*pb.Schema) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	for index, val := range schemas {
		if len(val.Schema) > 0 {
			schemas[index].Summary = schemaSummary(val.Schema)
		}
	}

	reqBody, err := json.Marshal(&pb.ModifySchemasRequest{ServiceId: serviceId, Schemas: schemas})
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiSchemasURL, project, serviceId),
		headers, reqBody)
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

func (c *SCClient) UpdateSchema(ctx context.Context, domainProject, serviceId string, schemaId string, schema string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.ModifySchemaRequest{
		ServiceId: serviceId,
		SchemaId:  schemaId,
		Schema:    schema,
		Summary:   schemaSummary(schema),
	})
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiSchemaURL, project, serviceId, schemaId),
		headers, reqBody)
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

func (c *SCClient) DeleteSchema(ctx context.Context, domainProject, serviceId string, schemaId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.DeleteSchemaRequest{ServiceId: serviceId, SchemaId: schemaId})
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiSchemaURL, project, serviceId, schemaId),
		headers, reqBody)
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

func schemaSummary(context string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(context)))
}
