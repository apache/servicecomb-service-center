package task

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	apt "github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	serviceUtil "github.com/apache/servicecomb-service-center/server/service/util"
)

// ClearNoInstanceService clears services which have no instance
// t means a service's existence time
func ClearNoInstanceServices(t time.Duration) error {
	services, err := serviceUtil.GetAllServicesAcrossDomainProject(context.Background())
	if err != nil {
		return err
	}
	if len(services) == 0 {
		log.Info("no service found, no need to clear")
		return nil
	}
	timeLimit := time.Now().Add(0 - t)
	log.Infof("clear no-instance services created before %s", timeLimit)
	timeLimitStamp := strconv.FormatInt(timeLimit.Unix(), 10)

	for domainProject, svcList := range services {
		if len(svcList) == 0 {
			continue
		}
		for _, svc := range svcList {
			//ignore a service if it is created after timeLimitStamp
			if svc.Timestamp > timeLimitStamp {
				continue
			}
			getInstsReq := &pb.GetInstancesRequest{
				ConsumerServiceId: svc.ServiceId,
				ProviderServiceId: svc.ServiceId,
			}
			svcCtxStr := "domainProject: " + domainProject + ", " +
				"env: " + svc.Environment + ", " +
				"service: " + util.StringJoin([]string{svc.AppId, svc.ServiceName, svc.Version}, apt.SPLIT)
			splitIndex := strings.Index(domainProject, apt.SPLIT)
			if splitIndex == -1 {
				log.Errorf(nil, "domain project format error: %s", domainProject)
				continue
			}
			domain := domainProject[:splitIndex]
			project := domainProject[splitIndex+1:]
			domainProjectCtx := util.SetDomainProject(context.Background(), domain, project)
			getInstsResp, err := apt.InstanceAPI.GetInstances(domainProjectCtx, getInstsReq)
			if err != nil {
				log.Errorf(err, "get instance failed, %s", svcCtxStr)
				continue
			}
			if getInstsResp.Response.GetCode() != pb.Response_SUCCESS {
				log.Errorf(nil, "get instance failed, %s, %s", getInstsResp.Response.GetMessage(), svcCtxStr)
				continue
			}
			//ignore a service if it has instances
			if len(getInstsResp.Instances) > 0 {
				continue
			}
			//delete this service
			delSvcReq := &pb.DeleteServiceRequest{
				ServiceId: svc.ServiceId,
				Force:     true,
			}
			delSvcResp, err := apt.ServiceAPI.Delete(domainProjectCtx, delSvcReq)
			if err != nil {
				log.Errorf(err, "clear service failed, %s", svcCtxStr)
				continue
			}

			if delSvcResp.Response.GetCode() != pb.Response_SUCCESS {
				log.Errorf(nil, "clear service failed, %s, %s", delSvcResp.Response.GetMessage(), svcCtxStr)
				continue
			}
			log.Warnf("clear service success, %s", svcCtxStr)
		}
	}
	return nil
}
