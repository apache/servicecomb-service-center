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
	"github.com/apache/incubator-servicecomb-service-center/server/admin/model"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"github.com/apache/incubator-servicecomb-service-center/version"
	"io/ioutil"
	"net/http"
)

const (
	apiVersionURL    = "/version"
	apiDumpURL       = "/v4/default/admin/dump"
	apiGetSchemasURL = "/v4/%s/registry/microservices/%s/schemas"
)

func (c *SCClient) commonHeaders() http.Header {
	var headers = make(http.Header)
	if len(Token) > 0 {
		headers.Set("X-Auth-Token", Token)
	}
	return headers
}

func (c *SCClient) GetScVersion() (*version.VersionSet, error) {
	resp, err := c.URLClient.HttpDo(http.MethodGet, Addr+apiVersionURL, c.commonHeaders(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(body))
	}

	v := &version.VersionSet{}
	err = json.Unmarshal(body, v)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	return v, nil
}

func (c *SCClient) GetScCache() (*model.Cache, error) {
	headers := c.commonHeaders()
	headers.Set("X-Domain-Name", "default")
	resp, err := c.URLClient.HttpDo(http.MethodGet, Addr+apiDumpURL, headers, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(body))
	}

	dump := &model.DumpResponse{}
	err = json.Unmarshal(body, dump)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	return dump.Cache, nil
}

func (c *SCClient) GetSchemasByServiceId(domainProject, serviceId string) ([]*pb.Schema, error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.commonHeaders()
	headers.Set("X-Domain-Name", domain)
	resp, err := c.URLClient.HttpDo(http.MethodGet,
		Addr+fmt.Sprintf(apiGetSchemasURL, project, serviceId)+"?withSchema=1",
		headers, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, string(body))
	}

	schemas := &pb.GetAllSchemaResponse{}
	err = json.Unmarshal(body, schemas)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}

	return schemas.Schemas, nil
}
