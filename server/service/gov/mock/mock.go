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

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
)

type Distributor struct {
	lbPolicies map[string]*gov.LoadBalancer
	name       string
}

func (d *Distributor) Create(kind, project string, spec []byte) error {
	p := &gov.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	log.Println(fmt.Sprintf("create %v", &p))
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return err
}
func (d *Distributor) Update(id, kind, project string, spec []byte) error {
	p := &gov.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	log.Println("update ", p)
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return err
}
func (d *Distributor) Delete(id, project string) error {
	delete(d.lbPolicies, id)
	return nil
}
func (d *Distributor) List(kind, project, app, env string) ([]byte, error) {
	r := make([]*gov.LoadBalancer, 0, len(d.lbPolicies))
	for _, g := range d.lbPolicies {
		r = append(r, g)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Get(id, project string) ([]byte, error) {
	return nil, nil
}
func (d *Distributor) Type() string {
	return svc.ConfigDistributorMock
}
func (d *Distributor) Name() string {
	return d.name
}
func new(opts config.DistributorOptions) (svc.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, lbPolicies: map[string]*gov.LoadBalancer{}}, nil
}
func init() {
	svc.InstallDistributor(svc.ConfigDistributorMock, new)
}
