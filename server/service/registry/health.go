package registry

import (
	"context"

	pb "github.com/go-chassis/cari/discovery"

	"github.com/apache/servicecomb-service-center/server/health"
)

func Readiness(_ context.Context) error {
	if err := health.GlobalReadinessChecker().Healthy(); err != nil {
		return pb.NewError(pb.ErrUnhealthy, err.Error())
	}
	return nil
}
