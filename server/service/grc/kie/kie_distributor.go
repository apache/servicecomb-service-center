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

package kie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/kie-client"

	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/config"
	grcsvc "github.com/apache/servicecomb-service-center/server/service/grc"
	rbacsvc "github.com/apache/servicecomb-service-center/server/service/rbac"
)

type Distributor struct {
	name   string
	client *kie.Client
}

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
	if kind == grcsvc.KindMatchGroup {
		err := d.generateID(ctx, project, p)
		if err != nil {
			return nil, err
		}
	}
	if kind == grcsvc.KindMatchGroup {
		setAliasIfEmpty(p.Spec, p.Name)
	}
	yamlByte, err := yaml.Marshal(p.Spec)
	if err != nil {
		return nil, err
	}
	kv := kie.KVRequest{
		Key:       toGovKeyPrefix(kind) + p.Name,
		Value:     string(yamlByte),
		Status:    grcsvc.StatusEnabled,
		ValueType: grcsvc.TypeText,
		Labels:    p.Selector,
	}
	res, err := d.client.Create(ctx, kv, kie.WithProject(project))
	if err != nil {
		log.Error("kie create failed", err)
		return nil, err
	}
	return []byte(res.ID), nil
}

func (d *Distributor) Update(ctx context.Context, kind, id, project string, p *gov.Policy) error {
	if kind == grcsvc.KindMatchGroup {
		setAliasIfEmpty(p.Spec, p.Name)
	}
	yamlByte, err := yaml.Marshal(p.Spec)
	if err != nil {
		return err
	}
	kv := kie.KVRequest{
		ID:     id,
		Value:  string(yamlByte),
		Status: p.Status,
	}
	_, err = d.client.Put(ctx, kv, kie.WithProject(project))
	if err != nil {
		log.Error("kie update failed", err)
		return err
	}
	return nil
}

func (d *Distributor) Delete(ctx context.Context, kind, id, project string) error {
	if kind == grcsvc.KindMatchGroup {
		// should remove all policies of this group
		return d.DeleteMatchGroup(ctx, id, project)
	}

	err := d.client.Delete(ctx, id, kie.WithProject(project))
	if err != nil {
		log.Error("kie delete failed", err)
		return err
	}
	return nil
}

func (d *Distributor) DeleteMatchGroup(ctx context.Context, id string, project string) error {
	policy, err := d.getPolicy(ctx, grcsvc.KindMatchGroup, id, project)
	if err != nil {
		log.Error("kie get failed", err)
		return err
	}

	ops := []kie.GetOption{
		kie.WithKey("wildcard(" + grcsvc.KeyPrefix + "*." + policy.Name + ")"),
		kie.WithLabels(policy.Selector),
		kie.WithRevision(0),
		kie.WithGetProject(project),
	}
	idList, _, err := d.client.List(ctx, ops...)
	if err != nil {
		log.Error("kie list failed", err)
		return err
	}
	var ids string
	for _, res := range idList.Data {
		ids += res.ID + ","
	}
	if len(ids) == 0 {
		return nil
	}

	err = d.client.Delete(ctx, ids[:len(ids)-1], kie.WithProject(project))
	if err != nil {
		log.Error("kie list failed", err)
		return err
	}
	return nil
}

func (d *Distributor) Display(ctx context.Context, project, app, env string) ([]byte, error) {
	list, _, err := d.listDataByKind(ctx, grcsvc.KindMatchGroup, project, app, env)
	if err != nil {
		return nil, err
	}
	policyMap := make(map[string]*gov.Policy)
	for _, kind := range PolicyNames {
		policies, _, err := d.listDataByKind(ctx, kind, project, app, env)
		if err != nil {
			continue
		}
		for _, policy := range policies.Data {
			item, err := d.transform(policy, kind)
			if err != nil {
				log.Warn(fmt.Sprintf("transform config failed: key is [%s], value is [%s]", policy.Key, policy.Value))
				continue
			}
			policyMap[item.Name+kind] = item
		}
	}
	r := make([]*gov.DisplayData, 0, list.Total)
	for _, item := range list.Data {
		match, err := d.transform(item, grcsvc.KindMatchGroup)
		if err != nil {
			log.Warn(fmt.Sprintf("transform config failed: key is [%s], value is [%s]", item.Key, item.Value))
			continue

		}
		var policies []*gov.Policy
		for _, kind := range PolicyNames {
			if policyMap[match.Name+kind] != nil {
				policies = append(policies, policyMap[match.Name+kind])
			}
		}
		result := &gov.DisplayData{
			Policies:   policies,
			MatchGroup: match,
		}
		r = append(r, result)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func setAliasIfEmpty(spec map[string]interface{}, name string) {
	if spec["alias"] == nil {
		spec["alias"] = name
		return
	}
	alias := spec["alias"].(string)
	if alias == "" {
		spec["alias"] = name
	}
}

func (d *Distributor) List(ctx context.Context, kind, project, app, env string) ([]byte, error) {
	list, _, err := d.listDataByKind(ctx, kind, project, app, env)
	if err != nil {
		return nil, err
	}
	r := make([]*gov.Policy, 0, list.Total)
	for _, item := range list.Data {
		policy, err := d.transform(item, kind)
		if err != nil {
			log.Warn(fmt.Sprintf("transform config failed: key is [%s], value is [%s]", item.Key, item.Value))
			continue
		}
		r = append(r, policy)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Get(ctx context.Context, kind, id, project string) ([]byte, error) {
	policy, err := d.getPolicy(ctx, kind, id, project)
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(policy, "", "  ")
	return b, nil
}

func (d *Distributor) getPolicy(ctx context.Context, kind string, id string, project string) (*gov.Policy, error) {
	kv, err := d.client.Get(ctx, id, kie.WithGetProject(project))
	if err != nil {
		return nil, err
	}
	policy, err := d.transform(kv, kind)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

func (d *Distributor) Type() string {
	return grcsvc.ConfigDistributorKie
}
func (d *Distributor) Name() string {
	return d.name
}

func initClient(endpoint string) *kie.Client {
	client, err := kie.NewClient(
		kie.Config{Endpoint: endpoint,
			DefaultLabels: map[string]string{},
			HTTPOptions: &httpclient.Options{
				SignRequest: rbacsvc.SignRequest,
			},
		})
	if err != nil {
		log.Fatal("init kie client failed, err: %s", err)
	}
	return client
}

func new(opts config.DistributorOptions) (grcsvc.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, client: initClient(opts.Endpoint)}, nil
}

func toSnake(name string) string {
	if name == "" {
		return ""
	}
	temp := strings.Split(name, "-")
	var buffer bytes.Buffer
	for num, v := range temp {
		vv := []rune(v)
		if num == 0 {
			buffer.WriteString(string(vv))
			continue
		}
		if len(vv) > 0 {
			if vv[0] >= 'a' && vv[0] <= 'z' { //首字母大写
				vv[0] -= 32
			}
			buffer.WriteString(string(vv))
		}
	}
	return buffer.String()
}

func (d *Distributor) listDataByKind(ctx context.Context, kind, project, app, env string) (*kie.KVResponse, int, error) {
	ops := []kie.GetOption{
		kie.WithKey("beginWith(" + toGovKeyPrefix(kind) + ")"),
		kie.WithRevision(0),
		kie.WithGetProject(project),
	}
	labels := map[string]string{}
	if env != grcsvc.EnvAll {
		labels[grcsvc.KeyEnvironment] = env
	}
	if app != "" {
		labels[grcsvc.KeyApp] = app
	}
	if len(labels) > 0 {
		ops = append(ops, kie.WithLabels(labels))
	}
	list, rev, err := d.client.List(ctx, ops...)
	if err == kie.ErrNoChanges {
		list = &kie.KVResponse{}
		err = nil
	}
	return list, rev, err
}

func (d *Distributor) generateID(ctx context.Context, project string, p *gov.Policy) error {
	if p.Name != "" {
		return nil
	}
	kind := grcsvc.KindMatchGroup
	list, _, err := d.listDataByKind(ctx, kind, project, p.Selector[grcsvc.KeyApp], p.Selector[grcsvc.KeyEnvironment])
	if err != nil {
		return err
	}
	var id string
	for {
		var repeat bool
		id = getID()
		govKey := toGovKeyPrefix(kind) + id
		for _, datum := range list.Data {
			if govKey == datum.Key {
				repeat = true
				break
			}
		}
		if !repeat {
			break
		}
	}
	p.Name = id
	return nil
}

func getID() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	b := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 4; i++ {
		result = append(result, b[r.Intn(len(b))])
	}
	return grcsvc.GroupNamePrefix + string(result)
}

func (d *Distributor) transform(kv *kie.KVDoc, kind string) (*gov.Policy, error) {
	goc := &gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Selector: gov.Selector{},
		},
	}
	spec := make(map[string]interface{})
	specJSON, _ := yaml.YAMLToJSON([]byte(kv.Value))
	err := json.Unmarshal(specJSON, &spec)
	if err != nil {
		log.Error("kie transform kv failed", err)
		return nil, err
	}
	goc.Kind = kind
	goc.ID = kv.ID
	goc.Status = kv.Status
	goc.Name = kv.Key[strings.LastIndex(kv.Key, ".")+1 : len(kv.Key)]
	goc.Spec = spec
	goc.Selector = kv.Labels
	goc.CreatTime = kv.CreatTime
	goc.UpdateTime = kv.UpdateTime
	return goc, nil
}

func toGovKeyPrefix(kind string) string {
	return grcsvc.KeyPrefix + toSnake(kind) + "."
}

func init() {
	grcsvc.InstallDistributor(grcsvc.ConfigDistributorKie, new)
}
