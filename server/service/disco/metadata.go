package disco

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/core"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/foundation/gopool"
)

func RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	//create service
	resp, err := registerService(ctx, request)
	if err != nil {
		return nil, err
	}

	if !hasServiceDetails(request) {
		return resp, nil
	}

	//create tag,rule,instances
	return registerServiceDetails(ctx, request, resp.ServiceId)
}

func registerService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if request == nil || request.Service == nil {
		log.Error(fmt.Sprintf("create micro-service failed: request body is empty, operator: %s", remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, "Request body is empty")
	}

	service := request.Service
	serviceFlag := util.StringJoin([]string{
		service.Environment, service.AppId, service.ServiceName, service.Version}, "/")
	datasource.SetServiceDefaultValue(service)

	if err := validator.ValidateCreateServiceRequest(request); err != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s", serviceFlag, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if quotaErr := checkServiceQuota(ctx); quotaErr != nil {
		log.Error(fmt.Sprintf("create micro-service[%s] failed, operator: %s", serviceFlag, remoteIP), quotaErr)
		return nil, quotaErr
	}

	return datasource.GetMetadataManager().RegisterService(ctx, request)
}

func registerServiceDetails(ctx context.Context, in *pb.CreateServiceRequest, serviceID string) (*pb.CreateServiceResponse, error) {
	var chanLen = 0
	errorsCh := make(chan error, 10)
	//create tags
	if in.Tags != nil && len(in.Tags) != 0 {
		chanLen++
		gopool.Go(func(_ context.Context) {
			req := &pb.AddServiceTagsRequest{
				ServiceId: serviceID,
				Tags:      in.Tags,
			}
			err := PutManyTags(ctx, req)
			errorsCh <- err
		})
	}
	// create instance
	if in.Instances != nil && len(in.Instances) != 0 {
		chanLen++
		gopool.Go(func(_ context.Context) {
			for _, ins := range in.Instances {
				req := &pb.RegisterInstanceRequest{
					Instance: ins,
				}
				req.Instance.ServiceId = serviceID
				_, err := RegisterInstance(ctx, req)
				errorsCh <- err
			}
		})
	}

	// handle result
	var errMessages []string
	for err := range errorsCh {
		chanLen--
		if err != nil {
			errMessages = append(errMessages, err.Error())
		}

		if 0 == chanLen {
			close(errorsCh)
		}
	}

	if len(errMessages) != 0 {
		return nil, pb.NewError(pb.ErrInvalidParams, fmt.Sprintf("errMessages: %v", errMessages))
	}

	log.Info(fmt.Sprintf("createServiceEx, serviceID: %s, operator: %s", serviceID, util.GetIPFromContext(ctx)))
	return &pb.CreateServiceResponse{
		ServiceId: serviceID,
	}, nil
}

func hasServiceDetails(in *pb.CreateServiceRequest) bool {
	if len(in.Rules) == 0 && len(in.Tags) == 0 && len(in.Instances) == 0 {
		return false
	}
	return true
}

func checkServiceQuota(ctx context.Context) error {
	if core.IsSCInstance(ctx) {
		log.Debug("skip quota check")
		return nil
	}
	return quotasvc.ApplyService(ctx, 1)
}

func UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateDeleteServiceRequest(request); err != nil {
		log.Error(fmt.Sprintf("delete micro-service[%s] failed, operator: %s", request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().UnregisterService(ctx, request)
}

func GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.MicroService, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetServiceRequest(in); err != nil {
		log.Error(fmt.Sprintf("get micro-service[%s] failed, operator: %s", in.ServiceId, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().GetService(ctx, in)
}

func ListService(ctx context.Context, in *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	resp, err := datasource.GetMetadataManager().ListService(ctx, in)
	if err == nil && len(resp.Services) > 0 {
		resp.Services = datasource.RemoveGlobalServices(in.WithShared, util.ParseDomainProject(ctx), resp.Services)
	}
	return resp, err
}

func FindService(ctx context.Context, in *pb.MicroServiceKey) (*pb.GetServicesResponse, error) {
	return datasource.GetMetadataManager().FindService(ctx, in)
}

func UnregisterManyService(ctx context.Context, request *pb.DelServicesRequest) (*pb.DelServicesResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	// 合法性检查
	if err := validator.ValidateUnregisterManyService(request); err != nil {
		log.Error(fmt.Sprintf("delete micro-services failed, operator: %s", remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	existFlag := map[string]bool{}
	nuoMultiCount := 0
	// 批量删除服务
	serviceRespChan := make(chan *pb.DelServicesRspInfo, len(request.ServiceIds))
	for _, serviceID := range request.ServiceIds {
		//ServiceId重复性检查
		if _, ok := existFlag[serviceID]; ok {
			log.Warn(fmt.Sprintf("duplicate micro-service[%s] serviceID", serviceID))
			continue
		} else {
			existFlag[serviceID] = true
			nuoMultiCount++
		}

		//执行删除服务操作
		gopool.Go(getDeleteServiceFunc(ctx, serviceID, request.Force, serviceRespChan))
	}

	//获取批量删除服务的结果
	count := 0
	responseCode := pb.ResponseSuccess
	delServiceRspInfo := make([]*pb.DelServicesRspInfo, 0, len(serviceRespChan))
	for serviceRespItem := range serviceRespChan {
		count++
		if len(serviceRespItem.ErrMessage) != 0 {
			responseCode = pb.ErrInvalidParams
		}
		delServiceRspInfo = append(delServiceRspInfo, serviceRespItem)
		//结果收集over，关闭通道
		if count == nuoMultiCount {
			close(serviceRespChan)
		}
	}

	log.Info(fmt.Sprintf("batch delete micro-services by serviceIDs[%d]: %v, result code: %d, operator: %s",
		len(request.ServiceIds), request.ServiceIds, responseCode, remoteIP))
	return &pb.DelServicesResponse{
		Services: delServiceRspInfo,
	}, nil
}

func getDeleteServiceFunc(ctx context.Context, serviceID string, force bool, serviceRespChan chan<- *pb.DelServicesRspInfo) func(context.Context) {
	return func(_ context.Context) {
		serviceRst := &pb.DelServicesRspInfo{
			ServiceId:  serviceID,
			ErrMessage: "",
		}
		err := UnregisterService(ctx, &pb.DeleteServiceRequest{
			ServiceId: serviceID,
			Force:     force,
		})
		if err != nil {
			serviceRst.ErrMessage = err.Error()
		}
		serviceRespChan <- serviceRst
	}
}

func ExistService(ctx context.Context, in *pb.GetExistenceRequest) (string, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateGetServiceExistenceRequest(in); err != nil {
		serviceFlag := util.StringJoin([]string{in.Environment, in.AppId, in.ServiceName, in.Version}, "/")
		log.Error(fmt.Sprintf("micro-service[%s] exist failed, operator: %s", serviceFlag, remoteIP), err)
		return "", pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().ExistService(ctx, in)
}

func PutServiceProperties(ctx context.Context, request *pb.UpdateServicePropsRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUpdateServicePropsRequest(request); err != nil {
		log.Error(fmt.Sprintf("update service[%s] properties failed, operator: %s",
			request.ServiceId, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().PutServiceProperties(ctx, request)
}

func ServiceUsage(ctx context.Context, request *pb.GetServiceCountRequest) (int64, error) {
	resp, err := datasource.GetMetadataManager().CountService(ctx, request)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
