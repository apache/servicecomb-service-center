// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-service-center/server/core/proto"
	"strconv"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

// ClearNoInstanceService clears services which have no instance
func ClearNoInstanceServices(serviceTTL time.Duration) error {
	services, err := serviceUtil.GetAllServicesAcrossDomainProject(context.Background())
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
		ctx, err := ctxFromDomainProject(domainProject)
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
				"service: " + util.StringJoin([]string{svc.AppId, svc.ServiceName, svc.Version}, apt.SPLIT)
			delSvcReq := &pb.DeleteServiceRequest{
				ServiceId: svc.ServiceId,
				Force:     true, //force delete
			}
			delSvcResp, err := apt.ServiceAPI.Delete(ctx, delSvcReq)
			if err != nil {
				log.Errorf(err, "clear service failed, %s", svcCtxStr)
				continue
			}
			if delSvcResp.Response.GetCode() != proto.Response_SUCCESS {
				log.Errorf(nil, "clear service failed, %s, %s", delSvcResp.Response.GetMessage(), svcCtxStr)
				continue
			}
			log.Warnf("clear service success, %s", svcCtxStr)
		}
	}
	return nil
}

func ctxFromDomainProject(domainProject string) (ctx context.Context, err error) {
	splitIndex := strings.Index(domainProject, apt.SPLIT)
	if splitIndex == -1 {
		return nil, errors.New("invalid domainProject: " + domainProject)
	}
	domain := domainProject[:splitIndex]
	project := domainProject[splitIndex+1:]
	return util.SetDomainProject(context.Background(), domain, project), nil
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
	getInstsResp, err := apt.InstanceAPI.GetInstances(ctx, getInstsReq)
	if err != nil {
		return false, err
	}
	if getInstsResp.Response.GetCode() != proto.Response_SUCCESS {
		return false, errors.New("get instance failed: " + getInstsResp.Response.GetMessage())
	}
	//ignore a service if it has instances
	if len(getInstsResp.Instances) > 0 {
		return false, nil
	}
	return true, nil
}
