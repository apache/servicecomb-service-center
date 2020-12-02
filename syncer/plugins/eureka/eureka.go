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

package eureka

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/apache/servicecomb-service-center/client"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const (
	PluginName = "eureka"

	apiApplications = "/apps"
)

func init() {
	// Register self as a repository plugin
	plugins.RegisterPlugin(&plugins.Plugin{
		Kind: plugins.PluginServicecenter,
		Name: PluginName,
		New:  New,
	})
}

type adaptor struct{}

func New() plugins.PluginInstance {
	return &adaptor{}
}

// New repository with endpoints
func (*adaptor) New(opts ...plugins.SCConfigOption) (plugins.Servicecenter, error) {
	cfg := plugins.ToSCConfig(opts...)
	client, err := client.NewLBClient(cfg.Endpoints, cfg.Merge())
	if err != nil {
		return nil, err
	}
	return &Client{LBClient: client, Cfg: cfg}, nil
}

type Client struct {
	*client.LBClient
	Cfg client.Config
}

// GetAll get and transform eureka data to SyncData
func (c *Client) GetAll(ctx context.Context) (*pb.SyncData, error) {
	method := http.MethodGet
	headers := c.CommonHeaders(method)
	resp, err := c.RestDoWithContext(ctx, method, apiApplications, headers, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	data := &eurekaData{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}

	return toSyncData(data), nil
}

// CommonHeaders Set the common header of the request
func (c *Client) CommonHeaders(method string) http.Header {
	var headers = make(http.Header)
	if len(c.Cfg.Token) > 0 {
		headers.Set("X-Auth-Token", c.Cfg.Token)
	}
	headers.Set("Accept", " application/json")
	headers.Set("Content-Type", "application/json")
	if method == http.MethodPut {
		headers.Set("x-netflix-discovery-replication", "true")
	}
	return headers
}

// toError response body to error
func (c *Client) toError(body []byte) error {
	return errors.New(util.BytesToStringWithNoCopy(body))
}
