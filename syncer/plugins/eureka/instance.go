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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const (
	apiInstances = "/apps/%s"
	apiInstance  = "/apps/%s/%s"
)

// RegisterInstance register instance to servicecenter
func (c *Client) RegisterInstance(ctx context.Context, domainProject, serviceID string, syncInstance *pb.SyncInstance) (string, error) {
	instance := toInstance(serviceID, syncInstance)
	method := http.MethodPost
	headers := c.CommonHeaders(method)
	body, err := json.Marshal(&InstanceRequest{Instance: instance})
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf(apiInstances, serviceID)
	resp, err := c.RestDoWithContext(ctx, method, apiURL, headers, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusNoContent {
		return "", c.toError(body)
	}

	return instance.InstanceID, nil
}

// UnregisterInstance unregister instance from servicecenter
func (c *Client) UnregisterInstance(ctx context.Context, domainProject, serviceID, instanceID string) error {
	method := http.MethodDelete
	headers := c.CommonHeaders(method)

	apiURL := fmt.Sprintf(apiInstance, serviceID, instanceID)
	resp, err := c.RestDoWithContext(ctx, method, apiURL, headers, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

// Heartbeat sends heartbeat to servicecenter
func (c *Client) Heartbeat(ctx context.Context, domainProject, serviceID, instanceID string) error {
	method := http.MethodPut
	headers := c.CommonHeaders(method)
	apiURL := fmt.Sprintf(apiInstance, serviceID, instanceID) + "?" + url.Values{"status": {UP}}.Encode()
	resp, err := c.RestDoWithContext(ctx, method, apiURL, headers, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}
