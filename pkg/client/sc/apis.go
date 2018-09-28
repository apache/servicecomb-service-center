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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"io/ioutil"
	"net/http"
)

const (
	apiVersionURL = "/version"
	apiDumpURL    = "/v4/default/admin/dump"
	apiSchemasURL = "/v4/%s/registry/microservices/%s/schemas"
	apiSchemaURL  = "/v4/%s/registry/microservices/%s/schemas/%s"
)

func (c *SCClient) toError(body []byte) *scerr.Error {
	message := new(scerr.Error)
	err := json.Unmarshal(body, message)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, util.BytesToStringWithNoCopy(body))
	}
	return message
}

func (c *SCClient) commonHeaders() http.Header {
	var headers = make(http.Header)
	if len(c.Config.Token) > 0 {
		headers.Set("X-Auth-Token", c.Config.Token)
	}
	return headers
}

func (c *SCClient) GetScVersion() (*version.VersionSet, *scerr.Error) {
	resp, err := c.URLClient.HttpDo(http.MethodGet, c.Config.Addr+apiVersionURL, c.commonHeaders(), nil)
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
		fmt.Println(string(body))
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return v, nil
}

func (c *SCClient) GetScCache() (*model.Cache, *scerr.Error) {
	headers := c.commonHeaders()
	// only default domain has admin permission
	headers.Set("X-Domain-Name", "default")
	resp, err := c.URLClient.HttpDo(http.MethodGet, c.Config.Addr+apiDumpURL, headers, nil)
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
		fmt.Println(string(body))
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return dump.Cache, nil
}

func (c *SCClient) GetSchemasByServiceId(domainProject, serviceId string) ([]*pb.Schema, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.commonHeaders()
	headers.Set("X-Domain-Name", domain)
	resp, err := c.URLClient.HttpDo(http.MethodGet,
		c.Config.Addr+fmt.Sprintf(apiSchemasURL, project, serviceId)+"?withSchema=1",
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
		fmt.Println(util.BytesToStringWithNoCopy(body))
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return schemas.Schemas, nil
}

func (c *SCClient) GetSchemaBySchemaId(domainProject, serviceId, schemaId string) (*pb.Schema, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.commonHeaders()
	headers.Set("X-Domain-Name", domain)
	resp, err := c.URLClient.HttpDo(http.MethodGet,
		c.Config.Addr+fmt.Sprintf(apiSchemaURL, project, serviceId, schemaId),
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
		fmt.Println(util.BytesToStringWithNoCopy(body))
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return &pb.Schema{
		SchemaId: schemaId,
		Schema:   schema.Schema,
		Summary:  schema.SchemaSummary,
	}, nil
}
