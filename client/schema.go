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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/pkg/errsvc"
)

const (
	apiSchemasURL = "/v4/%s/registry/microservices/%s/schemas"
	apiSchemaURL  = "/v4/%s/registry/microservices/%s/schemas/%s"
)

func (c *Client) CreateSchemas(ctx context.Context, domain, project, serviceID string, schemas []*pb.Schema) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	for index, val := range schemas {
		if len(val.Schema) > 0 {
			schemas[index].Summary = schemaSummary(val.Schema)
		}
	}

	reqBody, err := json.Marshal(&pb.ModifySchemasRequest{ServiceId: serviceID, Schemas: schemas})
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiSchemasURL, project, serviceID),
		headers, reqBody)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}
	return nil
}

func (c *Client) UpdateSchema(ctx context.Context, domain, project, serviceID string, schemaID string, schema string) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.ModifySchemaRequest{
		ServiceId: serviceID,
		SchemaId:  schemaID,
		Schema:    schema,
		Summary:   schemaSummary(schema),
	})
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiSchemaURL, project, serviceID, schemaID),
		headers, reqBody)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}
	return nil
}

func (c *Client) DeleteSchema(ctx context.Context, domain, project, serviceID string, schemaID string) *errsvc.Error {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.DeleteSchemaRequest{ServiceId: serviceID, SchemaId: schemaID})
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiSchemaURL, project, serviceID, schemaID),
		headers, reqBody)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}
	return nil
}

func (c *Client) GetSchemasByServiceID(ctx context.Context, domain, project, serviceID string) ([]*pb.Schema, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiSchemasURL, project, serviceID)+"?withSchema=1&"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	schemas := &pb.GetAllSchemaResponse{}
	err = json.Unmarshal(body, schemas)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return schemas.Schemas, nil
}

func (c *Client) GetSchemaBySchemaID(ctx context.Context, domain, project, serviceID, schemaID string) (*pb.Schema, *errsvc.Error) {
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiSchemaURL, project, serviceID, schemaID)+"?"+c.parseQuery(ctx),
		headers, nil)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	schema := &pb.GetSchemaResponse{}
	err = json.Unmarshal(body, schema)
	if err != nil {
		return nil, pb.NewError(pb.ErrInternal, err.Error())
	}

	return &pb.Schema{
		SchemaId: schemaID,
		Schema:   schema.Schema,
		Summary:  schema.SchemaSummary,
	}, nil
}

func schemaSummary(context string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(context)))
}
