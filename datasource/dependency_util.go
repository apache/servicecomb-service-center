package datasource

import (
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/go-chassis/cari/discovery"
)

func ParamsChecker(consumerInfo *pb.MicroServiceKey, providersInfo []*pb.MicroServiceKey) *pb.CreateDependenciesResponse {
	flag := make(map[string]bool, len(providersInfo))
	for _, providerInfo := range providersInfo {
		//存在带*的情况，后面的数据就不校验了
		if providerInfo.ServiceName == "*" {
			break
		}
		if len(providerInfo.AppId) == 0 {
			providerInfo.AppId = consumerInfo.AppId
		}

		version := providerInfo.Version
		if len(version) == 0 {
			return BadParamsResponse("Required provider version")
		}

		providerInfo.Version = ""
		if _, ok := flag[toString(providerInfo)]; ok {
			return BadParamsResponse("Invalid request body for provider info.Duplicate provider or (serviceName and appId is same).")
		}
		flag[toString(providerInfo)] = true
		providerInfo.Version = version
	}
	return nil
}

func BadParamsResponse(detailErr string) *pb.CreateDependenciesResponse {
	log.Errorf(nil, "request params is invalid. %s", detailErr)
	if len(detailErr) == 0 {
		detailErr = "Request params is invalid."
	}
	return &pb.CreateDependenciesResponse{
		Response: pb.CreateResponse(pb.ErrInvalidParams, detailErr),
	}
}

func toString(in *pb.MicroServiceKey) string {
	return path.GenerateProviderDependencyRuleKey(in.Tenant, in)
}
