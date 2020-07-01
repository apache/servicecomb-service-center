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
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/lb"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"net/http"
)

func NewLBClient(endpoints []string, options rest.URLClientOption) (*LBClient, error) {
	client, err := rest.GetURLClient(options)
	if err != nil {
		return nil, err
	}
	return &LBClient{
		Retries:   len(endpoints),
		LB:        lb.NewRoundRobinLB(endpoints),
		URLClient: client,
	}, nil
}

type LBClient struct {
	*rest.URLClient
	Retries int
	LB      lb.LoadBalancer
}

func (c *LBClient) Next() string {
	return c.LB.Next()
}

func (c *LBClient) RestDoWithContext(ctx context.Context, method string, api string, headers http.Header, body []byte) (resp *http.Response, err error) {
	var errs []string
	for i := 0; i < c.Retries; i++ {
		addr := c.Next()
		resp, err = c.HTTPDoWithContext(ctx, method, addr+api, headers, body)
		if err != nil {
			errs = append(errs, fmt.Sprintf("[%s]: %s", addr, err.Error()))
			continue
		}
		break
	}
	if err != nil {
		err = errors.New(util.StringJoin(errs, ", "))
	}
	return
}
