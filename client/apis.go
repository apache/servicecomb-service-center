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

	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/pkg/types"
	"github.com/apache/servicecomb-service-center/pkg/util"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/apache/servicecomb-service-center/version"
)

const (
	apiVersionURL  = "/version"
	apiDumpURL     = "/v4/default/admin/dump"
	apiClustersURL = "/v4/default/admin/clusters"
	apiHealthURL   = "/v4/default/registry/health"

	QueryGlobal util.CtxKey = "global"
)

func (c *Client) toError(body []byte) *scerr.Error {
	message := new(scerr.Error)
	err := json.Unmarshal(body, message)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, util.BytesToStringWithNoCopy(body))
	}
	return message
}

func (c *Client) parseQuery(ctx context.Context) (q string) {
	switch {
	case ctx.Value(QueryGlobal) == "1":
		q += "global=true"
	default:
		q += "global=false"
	}
	return
}

func (c *Client) GetScVersion(ctx context.Context) (*version.Set, *scerr.Error) {
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

	v := &version.Set{}
	err = json.Unmarshal(body, v)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return v, nil
}

func (c *Client) GetScCache(ctx context.Context) (*model.Cache, *scerr.Error) {
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

func (c *Client) GetClusters(ctx context.Context) (types.Clusters, *scerr.Error) {
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

func (c *Client) HealthCheck(ctx context.Context) *scerr.Error {
	headers := c.CommonHeaders(ctx)
	// only default domain has admin permission
	headers.Set("X-Domain-Name", "default")
	resp, err := c.RestDoWithContext(ctx, http.MethodGet, apiHealthURL, headers, nil)
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
