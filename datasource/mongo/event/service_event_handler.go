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

package event

import (
	"context"
	"fmt"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/mongo/dao"
	"github.com/apache/servicecomb-service-center/datasource/mongo/model"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
)

type ServiceEventHandler struct {
}

func NewServiceEventHandler() *ServiceEventHandler {
	return &ServiceEventHandler{}
}

func (h *ServiceEventHandler) Type() string {
	return model.ColumnService
}
func (h *ServiceEventHandler) OnEvent(evt sd.MongoEvent) {
	ms, ok := evt.Value.(model.Service)
	if !ok {
		log.Error("failed to assert service", datasource.ErrAssertFail)
		return
	}
	switch evt.Type {
	case pb.EVT_INIT, pb.EVT_CREATE:
		ctx := context.Background()
		err := newDomain(ctx, ms.Domain)
		if err != nil {
			log.Error(fmt.Sprintf("new domain %s failed", ms.Domain), err)
		}
		err = newProject(ctx, ms.Domain, ms.Project)
		if err != nil {
			log.Error(fmt.Sprintf("new project %s failed", ms.Project), err)
		}
	default:
	}
	if evt.Type == pb.EVT_INIT {
		return
	}

	log.Info(fmt.Sprintf("caught [%s] service[%s][%s/%s/%s/%s] event",
		evt.Type, ms.Service.ServiceId, ms.Service.Environment, ms.Service.AppId, ms.Service.ServiceName, ms.Service.Version))
}

func newDomain(ctx context.Context, domain string) error {
	filter := util.NewFilter(util.Domain(domain))
	exist, err := dao.ExistDomain(ctx, filter)
	if !exist && err == nil {
		err = dao.AddDomain(ctx, domain)
	}
	return err
}

func newProject(ctx context.Context, domain string, project string) error {
	filter := util.NewDomainProjectFilter(domain, project)
	exist, err := dao.ExistProject(ctx, filter)
	if !exist && err == nil {
		p := model.Project{
			Domain:  domain,
			Project: project,
		}
		err = dao.AddProject(ctx, p)
	}
	return err
}
