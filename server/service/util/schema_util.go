package util

import (
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"golang.org/x/net/context"
)

func CheckSchemaInfoExist(ctx context.Context, key string) (error, bool) {
	resp, errDo := store.Store().Schema().Search(ctx, &registry.PluginOp{
		Action:    registry.GET,
		Key:       []byte(key),
		CountOnly: true,
	})
	if errDo != nil {
		return errDo, false
	}
	if resp.Count == 0 {
		return nil, false
	}
	return nil, true
}
