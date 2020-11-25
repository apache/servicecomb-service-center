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
	"errors"
	"fmt"
	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/cluster"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/metrics"
	pb "github.com/go-chassis/cari/discovery"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"time"
)

func (ds *DataSource) SelfRegister(ctx context.Context) error {

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

func (ds *DataSource) registryService(pCtx context.Context) error {
	ctx := core.AddDefaultContextValue(pCtx)
	respE, err := datasource.Instance().ExistService(ctx, core.GetExistenceRequest())
	if err != nil {
		log.Error("query service center existence failed", err)
		return err
	}
	if respE.Response.GetCode() == pb.ResponseSuccess {
		log.Warnf("service center service[%s] already registered", respE.ServiceId)
		respG, err := datasource.Instance().GetService(ctx, core.GetServiceRequest(respE.ServiceId))
		if respG.Response.GetCode() != pb.ResponseSuccess {
			log.Errorf(err, "query service center service[%s] info failed", respE.ServiceId)
			return fmt.Errorf("service center service file lost")
		}
		core.Service = respG.Service
		return nil
	}

	respS, err := datasource.Instance().RegisterService(ctx, core.CreateServiceRequest())
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

	respI, err := datasource.Instance().RegisterInstance(ctx, core.RegisterInstanceRequest())
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

func (ds *DataSource) autoSelfHeartBeat() {
	//todo
}
func (ds *DataSource) SelfUnregister(ctx context.Context) error {

	if len(core.Instance.InstanceId) == 0 {
		return nil
	}

	ctx = core.AddDefaultContextValue(ctx)
	respI, err := datasource.Instance().UnregisterInstance(ctx, core.UnregisterInstanceRequest())
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

// OPS
func (ds *DataSource) ClearNoInstanceServices(ctx context.Context, ttl time.Duration) error {

	services, err := GetAllServicesAcrossDomainProject(ctx)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		log.Info("no service found, no need to clear")
		return nil
	}

	timeLimit := time.Now().Add(0 - ttl)
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
			svcCtxStr := "domainProject: " + domainProject + ", " +
				"env: " + svc.Environment + ", " +
				"service: " + util.StringJoin([]string{svc.AppId, svc.ServiceName, svc.Version}, path.SPLIT)
			delSvcReq := &pb.DeleteServiceRequest{
				ServiceId: svc.ServiceId,
				Force:     true, //force delete
			}
			delSvcResp, err := datasource.Instance().UnregisterService(ctx, delSvcReq)
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

func shouldClear(ctx context.Context, timeLimitStamp string, svc *pb.MicroService) (bool, error) {
	if svc.Timestamp > timeLimitStamp {
		return false, nil
	}

	getInstsReq := &pb.GetInstancesRequest{
		ConsumerServiceId: svc.ServiceId,
		ProviderServiceId: svc.ServiceId,
	}

	getInstsResp, err := datasource.Instance().GetInstances(ctx, getInstsReq)
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

func ctxFromDomainProject(pCtx context.Context, domainProject string) (ctx context.Context, err error) {
	splitIndex := strings.Index(domainProject, path.SPLIT)
	if splitIndex == -1 {
		return nil, errors.New("invalid domainProject: " + domainProject)
	}
	domain := domainProject[:splitIndex]
	project := domainProject[splitIndex+1:]
	return util.SetDomainProject(pCtx, domain, project), nil
}

func (ds *DataSource) UpgradeVersion(ctx context.Context) error {
	return nil
}

func (ds *DataSource) GetClusters(ctx context.Context) (cluster.Clusters, error) {
	return Configuration().Clusters, nil
}

func GetAllServicesAcrossDomainProject(ctx context.Context) (map[string][]*pb.MicroService, error) {
	domain := util.ParseDomain(ctx)
	project := util.ParseProject(ctx)

	filter := bson.M{"domain": domain, "project": project}

	findRes, err := client.GetMongoClient().Find(ctx, TBL_SERVICE, filter)
	if err != nil {
		return nil, err
	}

	services := make(map[string][]*pb.MicroService)

	for findRes.Next(ctx) {
		var mongoService MongoService
		err := findRes.Decode(&mongoService)
		if err != nil {
			return nil, err
		}
		domainProject := mongoService.Domain + "/" + mongoService.Project
		services[domainProject] = append(services[domainProject], mongoService.Service)
	}
	return services, nil
}
