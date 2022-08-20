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

package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofrs/uuid"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	grcsvc "github.com/apache/servicecomb-service-center/server/service/grc"
)

type Distributor struct {
	lbPolicies map[string]*gov.Policy
	name       string
}

const MatchGroup = "match-group"

var PolicyNames = []string{
	"retry",
	"rateLimiting",
	"circuitBreaker",
	"instanceIsolation",
	"faultInjection",
	"bulkhead",
	"loadbalance",
}

func (d *Distributor) Create(ctx context.Context, kind, project string, p *gov.Policy) ([]byte, error) {
	id, _ := uuid.NewV4()
	p.ID = id.String()
	p.Kind = kind
	log.Printf("create %v", &p)
	d.lbPolicies[p.GovernancePolicy.ID] = p
	return []byte(p.ID), nil
}

func (d *Distributor) Update(ctx context.Context, kind, id, project string, p *gov.Policy) error {
	if d.lbPolicies[id] == nil {
		return fmt.Errorf("id not exsit")
	}
	p.ID = id
	p.Kind = kind
	log.Println("update ", p)
	d.lbPolicies[p.GovernancePolicy.ID] = p
	return nil
}

func (d *Distributor) Delete(ctx context.Context, kind, id, project string) error {
	delete(d.lbPolicies, id)
	return nil
}

func (d *Distributor) Display(ctx context.Context, project, app, env string) ([]byte, error) {
	list := make([]*gov.Policy, 0)
	for _, g := range d.lbPolicies {
		if checkPolicy(g, MatchGroup, app, env) {
			list = append(list, g)
		}
	}
	policyMap := make(map[string]*gov.Policy)
	for _, g := range d.lbPolicies {
		for _, kind := range PolicyNames {
			if checkPolicy(g, kind, app, env) {
				policyMap[g.Name+kind] = g
			}
		}
	}
	r := make([]*gov.DisplayData, 0, len(list))
	for _, g := range list {
		policies := make([]*gov.Policy, 0)
		for _, kind := range PolicyNames {
			policies = append(policies, policyMap[g.Name+kind])
		}
		r = append(r, &gov.DisplayData{
			MatchGroup: g,
			Policies:   policies,
		})
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}
func (d *Distributor) List(ctx context.Context, kind, project, app, env string) ([]byte, error) {
	r := make([]*gov.Policy, 0, len(d.lbPolicies))
	for _, g := range d.lbPolicies {
		if checkPolicy(g, kind, app, env) {
			r = append(r, g)
		}
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func checkPolicy(g *gov.Policy, kind, app, env string) bool {
	return g.Kind == kind && g.Selector != nil && g.Selector["app"] == app && g.Selector["environment"] == env
}

func (d *Distributor) Get(ctx context.Context, kind, id, project string) ([]byte, error) {
	r := d.lbPolicies[id]
	if r == nil {
		return nil, nil
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Type() string {
	return grcsvc.ConfigDistributorMock
}
func (d *Distributor) Name() string {
	return d.name
}
func new(opts config.DistributorOptions) (grcsvc.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, lbPolicies: map[string]*gov.Policy{}}, nil
}
func init() {
	grcsvc.InstallDistributor(grcsvc.ConfigDistributorMock, new)
}
