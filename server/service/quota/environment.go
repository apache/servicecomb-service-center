package quota

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
)

const TypeEnvironment quota.ResourceType = "ENVIRONMENT"

func ApplyEnvironment(ctx context.Context, size int64) error {
	return quota.Apply(ctx, &quota.Request{
		QuotaType: TypeEnvironment,
		Domain:    util.ParseDomain(ctx),
		Project:   util.ParseProject(ctx),
		QuotaSize: size,
	})
}

func RemandEnvironment(ctx context.Context) {
	quota.Remand(ctx, TypeEnvironment)
}
