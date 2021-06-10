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

package mongo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client/model"
	mutil "github.com/apache/servicecomb-service-center/datasource/mongo/util"
	"github.com/apache/servicecomb-service-center/pkg/cluster"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/metrics"
)

type SCManager struct {
}

func (ds *SCManager) SelfRegister(ctx context.Context) error {
	err := ds.registryService(ctx)
	if err != nil {
		return err
	}

	// 实例信息
	err = ds.registryInstance(ctx)

	// wait heartbeat
	ds.autoSelfHeartBeat()

	metrics.ReportScInstance()
	return err
}
func (ds *SCManager) SelfUnregister(ctx context.Context) error {
	if len(core.Instance.InstanceId) == 0 {
		return nil
	}

	ctx = core.AddDefaultContextValue(ctx)
	respI, err := datasource.GetMetadataManager().UnregisterInstance(ctx, core.UnregisterInstanceRequest())
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

// OPS
func (ds *SCManager) ClearNoInstanceServices(ctx context.Context, ttl time.Duration) error {
	services, err := GetAllServicesAcrossDomainProject(ctx)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		log.Info("no service found, no need to clear")
		return nil
	}

	timeLimit := time.Now().Add(0 - ttl)
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
			svcCtxStr := "domainProject: " + domainProject + ", " +
				"env: " + svc.Environment + ", " +
				"service: " + util.StringJoin([]string{svc.AppId, svc.ServiceName, svc.Version}, path.SPLIT)
			delSvcReq := &pb.DeleteServiceRequest{
				ServiceId: svc.ServiceId,
				Force:     true, //force delete
			}
			delSvcResp, err := datasource.GetMetadataManager().UnregisterService(ctx, delSvcReq)
			if err != nil {
				log.Error(fmt.Sprintf("clear service failed, %s", svcCtxStr), err)
				continue
			}
			if delSvcResp.Response.GetCode() != pb.ResponseSuccess {
				log.Error(fmt.Sprintf("clear service failed %s %s", delSvcResp.Response.GetMessage(), svcCtxStr), err)
				continue
			}
			log.Warn(fmt.Sprintf("clear service success, %s", svcCtxStr))
		}
	}
	return nil
}

func (ds *SCManager) UpgradeVersion(ctx context.Context) error {
	return nil
}

func (ds *SCManager) GetClusters(ctx context.Context) (cluster.Clusters, error) {
	return nil, nil
}

func (ds *SCManager) registryService(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respE, err := datasource.GetMetadataManager().ExistService(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		return err
	}
	if respE.Response.GetCode() == pb.ResponseSuccess {
		log.Warn(fmt.Sprintf("service center service[%s] already registered", respE.ServiceId))
		respG, err := datasource.GetMetadataManager().GetService(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.GetCode() != pb.ResponseSuccess {
			log.Error(fmt.Sprintf("query service center service[%s] info failed", respE.ServiceId), err)
			return mutil.ErrLostServiceFile
		}
		core.Service = respG.Service
		return nil
	}

	respS, err := datasource.GetMetadataManager().RegisterService(ctx, core.CreateServiceRequest())
	if err != nil {
		log.Error("register service center failed", err)
		return err
	}
	core.Service.ServiceId = respS.ServiceId
	log.Info(fmt.Sprintf("register service center service[%s]", respS.ServiceId))
	return nil
}

func (ds *SCManager) registryInstance(pCtx context.Context) error {
	core.Instance.InstanceId = ""
	core.Instance.ServiceId = core.Service.ServiceId

	ctx := core.AddDefaultContextValue(pCtx)

	respI, err := datasource.GetMetadataManager().RegisterInstance(ctx, core.RegisterInstanceRequest())
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
	log.Info(fmt.Sprintf("register service center instance[%s/%s], endpoints is %s",
		core.Service.ServiceId, respI.InstanceId, core.Instance.Endpoints))
	return nil
}

func (ds *SCManager) selfHeartBeat(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respI, err := discosvc.Heartbeat(ctx, core.HeartbeatRequest())
	if err != nil {
		log.Error("send heartbeat failed", err)
		return err
	}
	if respI.Response.GetCode() == pb.ResponseSuccess {
		log.Debugf("update service center instance[%s/%s] heartbeat",
			core.Instance.ServiceId, core.Instance.InstanceId)
		return nil
	}
	err = fmt.Errorf(respI.Response.GetMessage())
	log.Errorf(err, "update service center instance[%s/%s] heartbeat failed",
		core.Instance.ServiceId, core.Instance.InstanceId)
	return err
}

func (ds *SCManager) autoSelfHeartBeat() {
	gopool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(core.Instance.HealthCheck.Interval) * time.Second):
				err := ds.selfHeartBeat(ctx)
				if err == nil {
					continue
				}
				//服务不存在，创建服务
				err = ds.SelfRegister(ctx)
				if err != nil {
					log.Errorf(err, "retry to register[%s/%s/%s/%s] failed",
						core.Service.Environment, core.Service.AppId, core.Service.ServiceName, core.Service.Version)
				}
			}
		}
	})
}

func GetAllServicesAcrossDomainProject(ctx context.Context) (map[string][]*pb.MicroService, error) {
	filter := mutil.NewBasicFilter(ctx)

	findRes, err := client.GetMongoClient().Find(ctx, model.CollectionService, filter)
	if err != nil {
		return nil, err
	}

	services := make(map[string][]*pb.MicroService)

	for findRes.Next(ctx) {
		var mongoService model.Service
		err := findRes.Decode(&mongoService)
		if err != nil {
			return nil, err
		}
		domainProject := mongoService.Domain + "/" + mongoService.Project
		services[domainProject] = append(services[domainProject], mongoService.Service)
	}
	return services, nil
}

func ctxFromDomainProject(pCtx context.Context, domainProject string) (ctx context.Context, err error) {
	splitIndex := strings.Index(domainProject, path.SPLIT)
	if splitIndex == -1 {
		return nil, mutil.NewError("invalid domainProject: ", domainProject)
	}
	domain := domainProject[:splitIndex]
	project := domainProject[splitIndex+1:]
	return util.SetDomainProject(pCtx, domain, project), nil
}

func shouldClear(ctx context.Context, timeLimitStamp string, svc *pb.MicroService) (bool, error) {
	if svc.Timestamp > timeLimitStamp {
		return false, nil
	}

	getInstsReq := &pb.GetInstancesRequest{
		ConsumerServiceId: svc.ServiceId,
		ProviderServiceId: svc.ServiceId,
	}

	getInstsResp, err := datasource.GetMetadataManager().GetInstances(ctx, getInstsReq)
	if err != nil {
		return false, err
	}
	if getInstsResp.Response.GetCode() != pb.ResponseSuccess {
		return false, mutil.NewError("get instance failed: ", getInstsResp.Response.GetMessage())
	}
	//ignore a service if it has instances
	if len(getInstsResp.Instances) > 0 {
		return false, nil
	}
	return true, nil
}
