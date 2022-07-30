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

package resource

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	kiemodel "github.com/apache/servicecomb-kie/pkg/model"
	kiecfg "github.com/apache/servicecomb-kie/server/config"
	kiedb "github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/plugin/security/tlsconf"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	"github.com/little-cui/etcdadpt"
)

const Config = "config"

func NewConfig(e *v1sync.Event) Resource {
	kind := config.GetString("registry.kind", "etcd", config.WithStandby("registry_plugin"))
	uri := config.GetString("registry.etcd.cluster.endpoints", "http://127.0.0.1:2379", config.WithStandby("manager_cluster"))
	isHTTPS := strings.Contains(strings.ToLower(uri), "https://")
	tlsOpts := tlsconf.GetOptions()
	kiecfg.Configurations.DB.Kind = kind
	err := kiedb.Init(kiecfg.DB{
		Kind:        kind,
		URI:         uri,
		SSLEnabled:  config.GetSSL().SslEnabled && isHTTPS,
		Timeout:     config.GetDuration("registry.etcd.request.timeout", etcdadpt.DefaultRequestTimeout, config.WithStandby("registry_timeout")).String(),
		RootCA:      filepath.Join(tlsOpts.Dir, "trust.cer"),
		CertFile:    filepath.Join(tlsOpts.Dir, "server.cer"),
		KeyFile:     filepath.Join(tlsOpts.Dir, "server_key.pem"),
		CertPwdFile: filepath.Join(tlsOpts.Dir, "cert_pwd"),
		VerifyPeer:  tlsOpts.VerifyPeer,
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("kie datasource[%s] init failed", kind), err)
	}
	c := &kvConfig{
		event: e,
	}
	c.resource = c
	return c
}

type kvConfig struct {
	defaultFailHandler

	event      *v1sync.Event
	input      *kiemodel.KVDoc
	cur        *kiemodel.KVDoc
	resource   docResource
	resourceID string
}

func (c *kvConfig) WithDomainProjectContext(ctx context.Context) context.Context {
	ctx = util.SetDomainProject(ctx,
		c.event.Opts[string(util.CtxDomain)],
		c.event.Opts[string(util.CtxProject)])
	return ctx
}

func (c *kvConfig) loadInput() error {
	c.input = new(kiemodel.KVDoc)
	callback := func() {
		c.resourceID = c.input.ID
	}
	param := newInputParam(c.input, callback)

	return newInputLoader(
		c.event,
		param,
		param,
		param,
	).loadInput()
}

func (c *kvConfig) LoadCurrentResource(ctx context.Context) *Result {
	err := c.loadInput()
	if err != nil {
		return FailResult(err)
	}

	ctx = c.WithDomainProjectContext(ctx)

	cur, err := c.resource.Get(ctx, c.resourceID)
	if err != nil {
		if errors.Is(err, kiedb.ErrKeyNotExists) {
			return nil
		}
		return FailResult(err)
	}
	c.cur = cur
	return nil
}

func (c *kvConfig) NeedOperate(ctx context.Context) *Result {
	ck := &checker{
		curNotNil: c.cur != nil,
		event:     c.event,
		updateTime: func() (int64, error) {
			return c.cur.UpdateTime, nil
		},
		resourceID: kiedb.TombstoneID(c.input),
	}
	ck.tombstoneLoader = ck
	return ck.needOperate(ctx)
}

func (c *kvConfig) Operate(ctx context.Context) *Result {
	ctx = c.WithDomainProjectContext(ctx)
	return newOperator(c).operate(ctx, c.event.Action)
}

type docResource interface {
	Create(ctx context.Context, doc *kiemodel.KVDoc) error
	Get(ctx context.Context, ID string) (*kiemodel.KVDoc, error)
	Update(ctx context.Context, doc *kiemodel.KVDoc) error
	Delete(ctx context.Context, ID string) error
}

func (c *kvConfig) Create(ctx context.Context, doc *kiemodel.KVDoc) error {
	_, err := kiedb.GetBroker().GetKVDao().Create(ctx, doc)
	return err
}

func (c *kvConfig) Get(ctx context.Context, ID string) (*kiemodel.KVDoc, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	return kiedb.GetBroker().GetKVDao().Get(ctx, &kiemodel.GetKVRequest{
		Project: project,
		Domain:  domain,
		ID:      ID,
	})
}

func (c *kvConfig) Update(ctx context.Context, doc *kiemodel.KVDoc) error {
	return kiedb.GetBroker().GetKVDao().Update(ctx, doc)
}

func (c *kvConfig) Delete(ctx context.Context, ID string) error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	_, err := kiedb.GetBroker().GetKVDao().FindOneAndDelete(ctx, ID, project, domain)
	return err
}

func (c *kvConfig) CreateHandle(ctx context.Context) error {
	if c.cur != nil {
		log.Warn(fmt.Sprintf("create config but doc id exist, %s", c.resourceID))
		return c.UpdateHandle(ctx)
	}
	return c.resource.Create(ctx, c.input)
}

func (c *kvConfig) UpdateHandle(ctx context.Context) error {
	if c.cur == nil {
		log.Warn(fmt.Sprintf("update action but account not exist, %s", c.resourceID))
		return c.CreateHandle(ctx)
	}
	return c.resource.Update(ctx, c.input)
}

func (c *kvConfig) DeleteHandle(ctx context.Context) error {
	return c.resource.Delete(ctx, c.input.ID)
}
