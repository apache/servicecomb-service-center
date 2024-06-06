package disco

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"

	"github.com/apache/servicecomb-service-center/datasource"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	quotasvc "github.com/apache/servicecomb-service-center/server/service/quota"
	"github.com/apache/servicecomb-service-center/server/service/validator"
)

var PreEnv util.CtxKey = "_pre_env"

func ListEnvironments(ctx context.Context) (*ev.GetEnvironmentsResponse, error) {
	return datasource.GetMetadataManager().ListEnvironments(ctx)
}

func RegistryEnvironment(ctx context.Context, request *ev.CreateEnvironmentRequest) (*ev.CreateEnvironmentResponse, error) {
	remoteIP := util.GetIPFromContext(ctx)

	if request == nil || request.Environment == nil {
		log.Error(fmt.Sprintf("create micro-service failed: request body is empty, operator: %s", remoteIP), nil)
		return nil, pb.NewError(pb.ErrInvalidParams, "Request body is empty")
	}

	env := request.Environment
	envFlag := util.StringJoin([]string{
		env.Name, env.Description}, "/")
	log.Info("will create environment:" + envFlag)

	if err := validator.ValidateCreateEnvironmentRequest(request); err != nil {
		log.Error(fmt.Sprintf("create environment[%s] failed, operator: %s", envFlag, remoteIP), err)
		return nil, pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	if quotaErr := checkEnvironmentQuota(ctx); quotaErr != nil {
		log.Error(fmt.Sprintf("create environment[%s] failed, operator: %s", envFlag, remoteIP), quotaErr)
		return nil, quotaErr
	}

	assignDefaultVal(env)

	return datasource.GetMetadataManager().RegisterEnvironment(ctx, request)
}

func assignDefaultVal(service *ev.Environment) {
	formatTenBase := 10
	service.Timestamp = strconv.FormatInt(time.Now().Unix(), formatTenBase)
	service.ModTimestamp = service.Timestamp
}

func GetEnvironment(ctx context.Context, in *ev.GetEnvironmentRequest) (*ev.Environment, error) {
	return datasource.GetMetadataManager().GetEnvironment(ctx, in)
}

func UpdateEnvironment(ctx context.Context, request *ev.UpdateEnvironmentRequest) error {
	remoteIP := util.GetIPFromContext(ctx)

	if err := validator.ValidateUpdateEnvironmentRequest(request); err != nil {
		log.Error(fmt.Sprintf("update environment[%s]  failed, operator: %s",
			request.Environment.ID, remoteIP), err)
		return pb.NewError(pb.ErrInvalidParams, err.Error())
	}

	return datasource.GetMetadataManager().UpdateEnvironment(ctx, request)
}

func UnRegistryEnvironment(ctx context.Context, request *ev.DeleteEnvironmentRequest) error {
	return datasource.GetMetadataManager().UnregisterEnvironment(ctx, request)
}

func checkEnvironmentQuota(ctx context.Context) error {
	b, _ := ctx.Value(PreEnv).(bool)
	if b {
		log.Debug("skip env quota check")
		return nil
	}
	return quotasvc.ApplyEnvironment(ctx, 1)
}

func EnvironmentUsage(ctx context.Context, request *ev.GetEnvironmentCountRequest) (int64, error) {
	resp, err := datasource.GetMetadataManager().CountEnvironment(ctx, request)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
