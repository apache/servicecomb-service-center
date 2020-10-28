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
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/server/core/config"
	"github.com/apache/servicecomb-service-center/server/service/gov"
	"log"
)

type Distributor struct {
	lbPolicies map[string]*model.LoadBalancer
	name       string
}

func (d *Distributor) Create(kind string, spec []byte) error {
	p := &model.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	log.Println(fmt.Sprintf("create %v", &p))
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return err
}
func (d *Distributor) Update(kind string, spec []byte) error {
	p := &model.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	log.Println("update ", p)
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return err
}
func (d *Distributor) Delete(name, kind string) error {
	delete(d.lbPolicies, name)
	return nil
}
func (d *Distributor) List(kind string) ([]byte, error) {
	r := make([]*model.LoadBalancer, len(d.lbPolicies))
	for _, g := range d.lbPolicies {
		r = append(r, g)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}
func (d *Distributor) Type() string {
	return gov.ConfigDistributorMock
}
func (d *Distributor) Name() string {
	return d.name
}
func new(opts config.DistributorOptions) (gov.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, lbPolicies: map[string]*model.LoadBalancer{}}, nil
}
func init() {
	gov.InstallDistributor(gov.ConfigDistributorMock, new)
}
