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
package admin_test

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/admin"
	"github.com/apache/servicecomb-service-center/server/admin/model"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	mgr "github.com/apache/servicecomb-service-center/server/plugin"
	"github.com/apache/servicecomb-service-center/server/plugin/discovery/etcd"
	etcd2 "github.com/apache/servicecomb-service-center/server/plugin/registry/etcd"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
	"github.com/astaxie/beego"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	beego.AppConfig.Set("registry_plugin", "etcd")
	mgr.RegisterPlugin(mgr.Plugin{mgr.REGISTRY, "etcd", etcd2.NewRegistry})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "buildin", etcd.NewRepository})
	mgr.RegisterPlugin(mgr.Plugin{mgr.DISCOVERY, "etcd", etcd.NewRepository})
}
func TestAdminService_Dump(t *testing.T) {
	t.Log("execute 'dump' operation,when get all,should be passed")
	resp, err := admin.AdminServiceAPI.Dump(getContext(), &model.DumpRequest{})
	assert.NoError(t, err)
	assert.Equal(t, pb.Response_SUCCESS, resp.Response.Code)
	t.Log("execute 'dump' operation,when get by domain project,should be passed")
	resp, err = admin.AdminServiceAPI.Dump(
		util.SetDomainProject(context.Background(), "x", "x"),
		&model.DumpRequest{})
	assert.NoError(t, err)
	assert.Equal(t, scerr.ErrForbidden, resp.Response.Code)
}

func getContext() context.Context {
	return util.SetContext(
		util.SetDomainProject(context.Background(), "default", "default"),
		serviceUtil.CTX_NOCACHE, "1")
}
