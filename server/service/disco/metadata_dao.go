package disco

import (
	"context"

	"github.com/apache/servicecomb-service-center/datasource"
	pb "github.com/go-chassis/cari/discovery"
)

func RegisterService(ctx context.Context, request *pb.CreateServiceRequest) (*pb.CreateServiceResponse, error) {
	return datasource.GetMetadataManager().RegisterService(ctx, request)
}
