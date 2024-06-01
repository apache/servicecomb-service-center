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

package registry

import (
	"context"
	"fmt"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/cari/env"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/foundation/gopool"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/server/service/sync"
)

func addDefaultContextValue(ctx context.Context) context.Context {
	ctx = util.WithNoCache(ctx)
	// set default domain/project
	ctx = util.SetDomainProject(ctx, datasource.RegistryDomain, datasource.RegistryProject)
	// register without quota check
	ctx = util.SetContext(ctx, core.CtxScSelf, true)
	// is a sync operation
	return sync.SetContext(ctx)
}

func SelfRegister(ctx context.Context) error {
	err := selfRegister(ctx)
	if err != nil {
		return err
	}
	// start send heart beat job
	autoSelfHeartBeat()
	return nil
}

func selfRegister(pCtx context.Context) error {
	ctx := addDefaultContextValue(pCtx)
	err := registerService(ctx)
	if err != nil {
		return err
	}
	// 实例信息
	return registerInstance(ctx)
}

func registerService(ctx context.Context) error {
	serviceID, err := discosvc.ExistService(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		if !errsvc.IsErrEqualCode(err, pb.ErrServiceNotExists) &&
			!errsvc.IsErrEqualCode(err, pb.ErrServiceVersionNotExists) {
			return err
		}
		return registerNewService(ctx)
	}

	log.Warn(fmt.Sprintf("service center service[%s] already registered", serviceID))
	core.Service, err = discosvc.GetService(ctx, core.GetServiceRequest(serviceID))
	if err != nil {
		log.Error(fmt.Sprintf("query service center service[%s] info failed", serviceID), err)
		return err
	}
	return nil
}

func registerNewService(ctx context.Context) error {
	respS, err := discosvc.RegisterService(ctx, core.CreateServiceRequest())
	if err != nil {
		log.Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	log.Info(fmt.Sprintf("register service center service[%s]", respS.ServiceId))
	return nil
}

func registerInstance(ctx context.Context) error {
	core.Instance.InstanceId = ""
	core.Instance.ServiceId = core.Service.ServiceId
	respI, err := discosvc.RegisterInstance(ctx, core.RegisterInstanceRequest())
	if err != nil {
		log.Error("register failed", err)
		return err
	}
	core.Instance.InstanceId = respI.InstanceId
	log.Info(fmt.Sprintf("register service center instance[%s/%s], endpoints is %s",
		core.Service.ServiceId, respI.InstanceId, core.Instance.Endpoints))
	return nil
}

func selfHeartBeat(pCtx context.Context) error {
	ctx := addDefaultContextValue(pCtx)
	err := discosvc.SendHeartbeat(ctx, core.HeartbeatRequest())
	if err != nil {
		log.Error("send heartbeat failed", err)
		return err
	}
	log.Debug(fmt.Sprintf("update service center instance[%s/%s] heartbeat",
		core.Instance.ServiceId, core.Instance.InstanceId))
	return nil
}

func autoSelfHeartBeat() {
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				err := selfHeartBeat(ctx)
				if err == nil {
					continue
				}
				// 服务不存在，创建服务
				err = selfRegister(ctx)
				if err != nil {
					log.Error(fmt.Sprintf("retry to register[%s/%s/%s/%s] failed",
						core.Service.Environment, core.Service.AppId, core.Service.ServiceName, core.Service.Version), err)
				}
			}
		}
	})
}

func SelfUnregister(pCtx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := addDefaultContextValue(pCtx)
	err := discosvc.UnregisterInstance(ctx, core.UnregisterInstanceRequest())
	if err != nil {
		log.Error(fmt.Sprintf("unregister service center instance[%s/%s] failed",
			core.Instance.ServiceId, core.Instance.InstanceId), err)
		return err
	}
	log.Warn(fmt.Sprintf("unregister service center instance[%s/%s]",
		core.Service.ServiceId, core.Instance.InstanceId))
	return nil
}

func SelfEnvRegister(ctx context.Context) error {
	err := selfEnvRegister(ctx)
	if err != nil {
		return err
	}
	return nil
}

func selfEnvRegister(pCtx context.Context) error {
	ctx := addDefaultEnvContextValue(pCtx)
	var req = new(env.CreateEnvironmentRequest)
	preEnv := env.Environment{Name: ""}
	req.Environment = &preEnv
	_, err := discosvc.RegistryEnvironment(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func addDefaultEnvContextValue(ctx context.Context) context.Context {
	// set default domain/project
	ctx = util.SetDomainProject(ctx, datasource.RegistryDomain, datasource.RegistryProject)
	ctx = util.SetContext(ctx, discosvc.PreEnv, true)
	ctx = util.SetContext(ctx, util.CtxRemoteIP, "127.0.0.1")
	return ctx
}
