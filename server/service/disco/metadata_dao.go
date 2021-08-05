package disco

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/server/service/validator"
	pb "github.com/go-chassis/cari/discovery"
)

func RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	return datasource.GetMetadataManager().RegisterService(ctx, request)
}

func UnregisterService(ctx context.Context, request *pb.DeleteServiceRequest) (*pb.DeleteServiceResponse, error) {
	return datasource.GetMetadataManager().UnregisterService(ctx, request)
}

func GetService(ctx context.Context, in *pb.GetServiceRequest) (*pb.MicroService, error) {
	err := validator.Validate(in)
	if err != nil {
		log.Error(fmt.Sprintf("get micro-service[%s] failed", in.ServiceId), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}
	return datasource.GetMetadataManager().GetService(ctx, in)
}
