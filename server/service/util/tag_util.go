package util

import (
	"encoding/json"
	apt "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	"github.com/ServiceComb/service-center/util"
	"golang.org/x/net/context"
)

func AddTagIntoETCD(ctx context.Context, tenant string, serviceId string, dataTags map[string]string) error {
	key := apt.GenerateServiceTagKey(tenant, serviceId)
	data, err := json.Marshal(dataTags)
	if err != nil {
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s:json marshal tag data failed.", serviceId)
		return err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, &registry.PluginOp{
		Action: registry.PUT,
		Key:    []byte(key),
		Value:  data,
	})
	if err != nil {
		util.LOGGER.Errorf(err, "add tag into etcd,serviceId %s: commit tag data into etcd failed.", serviceId)
		return err
	}
	return nil
}

func GetTagsUtils(ctx context.Context, tenant string, serviceId string) (map[string]string, error) {
	tags := map[string]string{}

	key := apt.GenerateServiceTagKey(tenant, serviceId)
	resp, err := store.Store().ServiceTag().Search(ctx, &registry.PluginOp{
		Action: registry.GET,
		Key:    []byte(key),
	})
	if err != nil {
		return tags, err
	}
	if len(resp.Kvs) != 0 {
		util.LOGGER.Debugf("start unmarshal service tags file: %s", key)
		err = json.Unmarshal(resp.Kvs[0].Value, &tags)
		if err != nil {
			return tags, err
		}
	}
	return tags, nil
}
