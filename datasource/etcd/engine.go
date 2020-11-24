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
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/cluster"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"strconv"
	"strings"
	"time"

	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/go-chassis/cari/discovery"
)

func (ds *DataSource) SelfRegister(ctx context.Context) error {
	err := ds.registryService(ctx)
	if err != nil {
		return err
	}
	// 实例信息
	err = ds.registryInstance(ctx)
	// start send heart beat job
	ds.autoSelfHeartBeat()
	// report the metrics
	metrics.ReportScInstance()
	return err
}

func (ds *DataSource) registryService(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respE, err := core.ServiceAPI.Exist(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		return err
	}
	if respE.Response.GetCode() == pb.ResponseSuccess {
		log.Warnf("service center service[%s] already registered", respE.ServiceId)
		respG, err := core.ServiceAPI.GetOne(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.GetCode() != pb.ResponseSuccess {
			log.Errorf(err, "query service center service[%s] info failed", respE.ServiceId)
			return fmt.Errorf("service center service file lost")
		}
		core.Service = respG.Service
		return nil
	}

	respS, err := core.ServiceAPI.Create(ctx, core.CreateServiceRequest())
	if err != nil {
		log.Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	log.Infof("register service center service[%s]", respS.ServiceId)
	return nil
}

func (ds *DataSource) registryInstance(pCtx context.Context) error {
	core.Instance.InstanceId = ""
	core.Instance.ServiceId = core.Service.ServiceId

	ctx := core.AddDefaultContextValue(pCtx)

	respI, err := core.InstanceAPI.Register(ctx, core.RegisterInstanceRequest())
	if err != nil {
		log.Error("register failed", err)
		return err
	}
	if respI.Response.GetCode() != pb.ResponseSuccess {
		err = fmt.Errorf("register service center[%s] instance failed, %s",
			core.Instance.ServiceId, respI.Response.GetMessage())
		log.Error(err.Error(), nil)
		return err
	}
	core.Instance.InstanceId = respI.InstanceId
	log.Infof("register service center instance[%s/%s], endpoints is %s",
		core.Service.ServiceId, respI.InstanceId, core.Instance.Endpoints)
	return nil
}

func (ds *DataSource) SelfUnregister(pCtx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Unregister(ctx, core.UnregisterInstanceRequest())
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
	log.Warnf("unregister service center instance[%s/%s]",
		core.Service.ServiceId, core.Instance.InstanceId)
	return nil
}

func (ds *DataSource) selfHeartBeat(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := core.InstanceAPI.Heartbeat(ctx, core.HeartbeatRequest())
	if err != nil {
		log.Error("sen heartbeat failed", err)
		return err
	}
	if respI.Response.GetCode() == pb.ResponseSuccess {
		log.Debugf("update service center instance[%s/%s] heartbeat",
			core.Instance.ServiceId, core.Instance.InstanceId)
		return nil
	}
	log.Errorf(err, "update service center instance[%s/%s] heartbeat failed",
		core.Instance.ServiceId, core.Instance.InstanceId)

	//服务不存在，创建服务
	err = ds.SelfRegister(pCtx)
	if err != nil {
		log.Errorf(err, "retry to register[%s/%s/%s/%s] failed",
			core.Service.Environment, core.Service.AppId, core.Service.ServiceName, core.Service.Version)
	}
	return err
}

func (ds *DataSource) autoSelfHeartBeat() {
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				_ = ds.selfHeartBeat(ctx)
			}
		}
	})
}

// ClearNoInstanceService clears services which have no instance
func (ds *DataSource) ClearNoInstanceServices(ctx context.Context, serviceTTL time.Duration) error {
	services, err := serviceUtil.GetAllServicesAcrossDomainProject(ctx)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		log.Info("no service found, no need to clear")
		return nil
	}
	timeLimit := time.Now().Add(0 - serviceTTL)
	log.Infof("clear no-instance services created before %s", timeLimit)
	timeLimitStamp := strconv.FormatInt(timeLimit.Unix(), 10)

	for domainProject, svcList := range services {
		if len(svcList) == 0 {
			continue
		}
		ctx, err := ctxFromDomainProject(ctx, domainProject)
		if err != nil {
			log.Errorf(err, "get domain project context failed")
			continue
		}
		for _, svc := range svcList {
			if svc == nil {
				continue
			}
			ok, err := shouldClear(ctx, timeLimitStamp, svc)
			if err != nil {
				log.Errorf(err, "check service clear necessity failed")
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
				log.Errorf(err, "clear service failed, %s", svcCtxStr)
				continue
			}
			if delSvcResp.Response.GetCode() != pb.ResponseSuccess {
				log.Errorf(nil, "clear service failed, %s, %s", delSvcResp.Response.GetMessage(), svcCtxStr)
				continue
			}
			log.Warnf("clear service success, %s", svcCtxStr)
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
	getInstsResp, err := core.InstanceAPI.GetInstances(ctx, getInstsReq)
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

func (ds *DataSource) GetClusters(ctx context.Context) (cluster.Clusters, error) {
	return Configuration().Clusters, nil
}
