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

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/mux"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/core"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
	"github.com/apache/servicecomb-service-center/version"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
	"github.com/little-cui/etcdadpt"
)

type SCManager struct {
}

func (sm *SCManager) SelfRegister(ctx context.Context) error {
	err := sm.selfRegister(ctx)
	if err != nil {
		return err
	}
	// start send heart beat job
	sm.autoSelfHeartBeat()
	return nil
}

func (sm *SCManager) selfRegister(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	err := sm.registerService(ctx)
	if err != nil {
		return err
	}
	// 实例信息
	return sm.registerInstance(ctx)
}

func (sm *SCManager) registerService(ctx context.Context) error {
	respE, err := core.ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		return err
	}
	if respE.Response.GetCode() == pb.ResponseSuccess {
		log.Warn(fmt.Sprintf("service center service[%s] already registered", respE.ServiceId))
		respG, err := core.ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.GetCode() != pb.ResponseSuccess {
			log.Error(fmt.Sprintf("query service center service[%s] info failed", respE.ServiceId), err)
			return datasource.ErrServiceNotExists
		}
		core.Service = respG.Service
		return nil
	}

	respS, err := core.ServiceAPI.Create(ctx, core.CreateServiceRequest())
	if err != nil {
		log.Error("register service center failed", err)
		return err
	}
	if respS.Response.GetCode() != pb.ResponseSuccess {
		log.Error("register service center failed, msg: "+respS.Response.GetMessage(), nil)
		return errors.New(respS.Response.GetMessage())
	}
	core.Service.ServiceId = respS.ServiceId
	log.Info(fmt.Sprintf("register service center service[%s]", respS.ServiceId))
	return nil
}

func (sm *SCManager) registerInstance(ctx context.Context) error {
	core.Instance.InstanceId = ""
	core.Instance.ServiceId = core.Service.ServiceId
	respI, err := discosvc.RegisterInstance(ctx, core.RegisterInstanceRequest())
	if err != nil {
		log.Error("register failed", err)
		return err
	}
	if respI.Response.GetCode() != pb.ResponseSuccess {
		log.Error(fmt.Sprintf("register service center[%s] instance failed, %s",
			core.Instance.ServiceId, respI.Response.GetMessage()), nil)
		return errors.New(respI.Response.GetMessage())
	}
	core.Instance.InstanceId = respI.InstanceId
	log.Info(fmt.Sprintf("register service center instance[%s/%s], endpoints is %s",
		core.Service.ServiceId, respI.InstanceId, core.Instance.Endpoints))
	return nil
}

func (sm *SCManager) selfHeartBeat(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := discosvc.Heartbeat(ctx, core.HeartbeatRequest())
	if err != nil {
		log.Error("send heartbeat failed", err)
		return err
	}
	if respI.Response.GetCode() == pb.ResponseSuccess {
		log.Debug(fmt.Sprintf("update service center instance[%s/%s] heartbeat",
			core.Instance.ServiceId, core.Instance.InstanceId))
		return nil
	}
	err = fmt.Errorf(respI.Response.GetMessage())
	log.Error(fmt.Sprintf("update service center instance[%s/%s] heartbeat failed",
		core.Instance.ServiceId, core.Instance.InstanceId), err)
	return err
}

func (sm *SCManager) autoSelfHeartBeat() {
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				err := sm.selfHeartBeat(ctx)
				if err == nil {
					continue
				}
				//服务不存在，创建服务
				err = sm.selfRegister(ctx)
				if err != nil {
					log.Error(fmt.Sprintf("retry to register[%s/%s/%s/%s] failed",
						core.Service.Environment, core.Service.AppId, core.Service.ServiceName, core.Service.Version), err)
				}
			}
		}
	})
}

func (sm *SCManager) SelfUnregister(pCtx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := discosvc.UnregisterInstance(ctx, core.UnregisterInstanceRequest())
	if err != nil {
		log.Error("unregister failed", err)
		return err
	}
	if respI.Response.GetCode() != pb.ResponseSuccess {
		err = fmt.Errorf("unregister service center instance[%s/%s] failed, %s",
			core.Instance.ServiceId, core.Instance.InstanceId, respI.Response.GetMessage())
		log.Error(err.Error(), nil)
		return err
	}
	log.Warn(fmt.Sprintf("unregister service center instance[%s/%s]",
		core.Service.ServiceId, core.Instance.InstanceId))
	return nil
}

// ClearNoInstanceService clears services which have no instance
func (sm *SCManager) ClearNoInstanceServices(ctx context.Context, serviceTTL time.Duration) error {
	services, err := serviceUtil.GetAllServicesAcrossDomainProject(ctx)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		log.Info("no service found, no need to clear")
		return nil
	}
	timeLimit := time.Now().Add(0 - serviceTTL)
	log.Info(fmt.Sprintf("clear no-instance services created before %s", timeLimit))
	timeLimitStamp := strconv.FormatInt(timeLimit.Unix(), 10)

	for domainProject, svcList := range services {
		if len(svcList) == 0 {
			continue
		}
		ctx, err := ctxFromDomainProject(ctx, domainProject)
		if err != nil {
			log.Error("get domain project context failed", err)
			continue
		}
		for _, svc := range svcList {
			if svc == nil {
				continue
			}
			ok, err := shouldClear(ctx, timeLimitStamp, svc)
			if err != nil {
				log.Error("check service clear necessity failed", err)
				continue
			}
			if !ok {
				continue
			}
			//delete this service
			svcCtxStr := "domainProject: " + domainProject + ", " +
				"env: " + svc.Environment + ", " +
				"service: " + util.StringJoin([]string{svc.AppId, svc.ServiceName, svc.Version}, path.SPLIT)
			delSvcReq := &pb.DeleteServiceRequest{
				ServiceId: svc.ServiceId,
				Force:     true, //force delete
			}
			delSvcResp, err := core.ServiceAPI.Delete(ctx, delSvcReq)
			if err != nil {
				log.Error(fmt.Sprintf("clear service failed, %s", svcCtxStr), err)
				continue
			}
			if delSvcResp.Response.GetCode() != pb.ResponseSuccess {
				log.Error(fmt.Sprintf("clear service failed, %s, %s", delSvcResp.Response.GetMessage(), svcCtxStr), err)
				continue
			}
			log.Warn(fmt.Sprintf("clear service success, %s", svcCtxStr))
		}
	}
	return nil
}

func ctxFromDomainProject(pCtx context.Context, domainProject string) (ctx context.Context, err error) {
	splitIndex := strings.Index(domainProject, path.SPLIT)
	if splitIndex == -1 {
		return nil, errors.New("invalid domainProject: " + domainProject)
	}
	domain := domainProject[:splitIndex]
	project := domainProject[splitIndex+1:]
	return util.SetDomainProject(pCtx, domain, project), nil
}

//check whether a service should be cleared
func shouldClear(ctx context.Context, timeLimitStamp string, svc *pb.MicroService) (bool, error) {
	//ignore a service if it is created after timeLimitStamp
	if svc.Timestamp > timeLimitStamp {
		return false, nil
	}
	getInstsReq := &pb.GetInstancesRequest{
		ConsumerServiceId: svc.ServiceId,
		ProviderServiceId: svc.ServiceId,
	}
	getInstsResp, err := discosvc.GetInstances(ctx, getInstsReq)
	if err != nil {
		return false, err
	}
	if getInstsResp.Response.GetCode() != pb.ResponseSuccess {
		return false, errors.New("get instance failed: " + getInstsResp.Response.GetMessage())
	}
	//ignore a service if it has instances
	if len(getInstsResp.Instances) > 0 {
		return false, nil
	}
	return true, nil
}

func (sm *SCManager) GetClusters(ctx context.Context) (etcdadpt.Clusters, error) {
	return etcdadpt.ListCluster(ctx)
}
func (sm *SCManager) UpgradeServerVersion(ctx context.Context) error {
	bytes, err := json.Marshal(config.Server)
	if err != nil {
		return err
	}
	return etcdadpt.PutBytes(ctx, path.GetServerInfoKey(), bytes)
}
func (sm *SCManager) UpgradeVersion(ctx context.Context) error {
	lock, err := mux.Lock(mux.GlobalLock)

	if err != nil {
		log.Error("wait for server ready failed", err)
		return err
	}
	if needUpgrade(ctx) {
		config.Server.Version = version.Ver().Version

		if err := sm.UpgradeServerVersion(ctx); err != nil {
			log.Error("upgrade server version failed", err)
			os.Exit(1)
		}
	}
	err = lock.Unlock()
	if err != nil {
		log.Error("", err)
	}
	return err
}
