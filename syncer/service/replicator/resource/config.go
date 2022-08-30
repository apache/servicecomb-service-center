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

	kiemodel "github.com/apache/servicecomb-kie/pkg/model"
	kiedb "github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
)

const Config = "config"

func NewConfig(e *v1sync.Event) Resource {
	kind := config.GetString("registry.kind", "etcd", config.WithStandby("registry_plugin"))
	err := kiedb.Init(kind)
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
	revision, err := kiedb.GetBroker().GetRevisionDao().ApplyRevision(ctx, doc.Domain)
	if err != nil {
		return fmt.Errorf("apply kv revision failed, %s", err.Error())
	}
	completeKV(doc, revision)
	doc, err = kiedb.GetBroker().GetKVDao().Create(ctx, doc)
	if err != nil {
		return fmt.Errorf("create kv failed, %s", err.Error())
	}
	err = kiedb.GetBroker().GetHistoryDao().AddHistory(ctx, doc)
	if err != nil {
		log.Warn(fmt.Sprintf("can not updateKeyValue version for [%s] [%s] in [%s], err: %s",
			doc.Key, doc.Labels, doc.Domain, err))
	}
	return nil
}

func completeKV(kv *kiemodel.KVDoc, revision int64) {
	kv.UpdateRevision = revision
	kv.CreateRevision = revision
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
	var err error
	doc.UpdateRevision, err = kiedb.GetBroker().GetRevisionDao().ApplyRevision(ctx, doc.Domain)
	if err != nil {
		return fmt.Errorf("apply kv revision failed, %s", err.Error())
	}
	err = kiedb.GetBroker().GetKVDao().Update(ctx, doc)
	if err != nil {
		return fmt.Errorf("update kv failed, %s", err.Error())
	}
	err = kiedb.GetBroker().GetHistoryDao().AddHistory(ctx, doc)
	if err != nil {
		log.Warn(fmt.Sprintf("can not add revision for [%s] [%s] in [%s], err: %s",
			doc.Key, doc.Labels, doc.Domain, err.Error()))
	}
	return nil
}

func (c *kvConfig) Delete(ctx context.Context, ID string) error {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)
	_, err := kiedb.GetBroker().GetKVDao().FindOneAndDelete(ctx, ID, project, domain)
	if err != nil {
		return fmt.Errorf("delete kv failed, %s", err.Error())
	}
	_, err = kiedb.GetBroker().GetRevisionDao().ApplyRevision(ctx, domain)
	if err != nil {
		return fmt.Errorf("the kv [%s] is deleted, but increase revision failed, %s", ID, err.Error())
	}
	err = kiedb.GetBroker().GetHistoryDao().DelayDeletionTime(ctx, []string{ID}, project, domain)
	if err != nil {
		log.Warn(fmt.Sprintf("add delete time to [%s] failed, err: %s", ID, err.Error()))
	}
	return nil
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
