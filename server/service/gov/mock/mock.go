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
	"encoding/json"
	"fmt"
	"log"

	uuid "github.com/satori/go.uuid"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
)

type Distributor struct {
	lbPolicies map[string]*gov.Policy
	name       string
}

const MatchGroup = "match-group"

var PolicyNames = []string{"retry", "rateLimiting", "circuitBreaker", "bulkhead"}

func (d *Distributor) Create(kind, project string, spec []byte) ([]byte, error) {
	p := &gov.Policy{}
	err := json.Unmarshal(spec, p)
	p.ID = uuid.NewV4().String()
	p.Kind = kind
	log.Println(fmt.Sprintf("create %v", &p))
	d.lbPolicies[p.GovernancePolicy.ID] = p
	return []byte(p.ID), err
}

func (d *Distributor) Update(id, kind, project string, spec []byte) error {
	if d.lbPolicies[id] == nil {
		return fmt.Errorf("id not exsit")
	}
	p := &gov.Policy{}
	err := json.Unmarshal(spec, p)
	p.ID = id
	p.Kind = kind
	log.Println("update ", p)
	d.lbPolicies[p.GovernancePolicy.ID] = p
	return err
}

func (d *Distributor) Delete(id, project string) error {
	delete(d.lbPolicies, id)
	return nil
}

func (d *Distributor) Display(project, app, env string) ([]byte, error) {
	list := make([]*gov.Policy, 0)
	for _, g := range d.lbPolicies {
		if g.Kind == MatchGroup && g.Selector.App == app && g.Selector.Environment == env {
			list = append(list, g)
		}
	}
	policyMap := make(map[string]*gov.Policy)
	for _, g := range d.lbPolicies {
		for _, kind := range PolicyNames {
			if g.Kind == kind && g.Selector.App == app && g.Selector.Environment == env {
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
func (d *Distributor) List(kind, project, app, env string) ([]byte, error) {
	r := make([]*gov.Policy, 0, len(d.lbPolicies))
	for _, g := range d.lbPolicies {
		if g.Kind == kind && g.Selector.App == app && g.Selector.Environment == env {
			r = append(r, g)
		}
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Get(kind, id, project string) ([]byte, error) {
	r := d.lbPolicies[id]
	if r == nil {
		return nil, nil
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Type() string {
	return svc.ConfigDistributorMock
}
func (d *Distributor) Name() string {
	return d.name
}
func new(opts config.DistributorOptions) (svc.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, lbPolicies: map[string]*gov.Policy{}}, nil
}
func init() {
	svc.InstallDistributor(svc.ConfigDistributorMock, new)
}
