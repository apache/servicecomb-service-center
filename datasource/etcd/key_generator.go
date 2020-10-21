package etcd

import (
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"strings"
)

const (
	RegistryRootKey      = "cse-sr"
	RegistryProjectKey   = "projects"
	SPLIT                = "/"
	RegistryDomainKey    = "domains"
	RegistryServiceKey   = "ms"
	RegistryIndex        = "indexes"
	RegistryDepsQueueKey = "dep-queue"
)

func GetRootKey() string {
	return SPLIT + RegistryRootKey
}

func (ds *DataSource) GenerateAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}

func GenerateETCDAccountKey(name string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		"accounts",
		name,
	}, SPLIT)
}

func GenerateETCDProjectKey(domain, project string) string {
	return util.StringJoin([]string{
		GetProjectRootKey(domain),
		project,
	}, SPLIT)
}

func GetProjectRootKey(domain string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryProjectKey,
		domain,
	}, SPLIT)
}

func GenerateETCDDomainKey(domain string) string {
	return util.StringJoin([]string{
		GetDomainRootKey(),
		domain,
	}, SPLIT)
}

func GetDomainRootKey() string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryDomainKey,
	}, SPLIT)
}

func GenerateServiceIndexKey(key *registry.MicroServiceKey) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(key.Tenant),
		key.Environment,
		key.AppId,
		key.ServiceName,
		key.Version,
	}, SPLIT)
}

func GetServiceIndexRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryIndex,
		domainProject,
	}, SPLIT)
}

func KvToResponse(key []byte) (keys []string) {
	return strings.Split(util.BytesToStringWithNoCopy(key), SPLIT)
}

func GetInfoFromSvcIndexKV(key []byte) *registry.MicroServiceKey {
	keys := KvToResponse(key)
	l := len(keys)
	if l < 6 {
		return nil
	}
	domainProject := fmt.Sprintf("%s/%s", keys[l-6], keys[l-5])
	return &registry.MicroServiceKey{
		Tenant:      domainProject,
		Environment: keys[l-4],
		AppId:       keys[l-3],
		ServiceName: keys[l-2],
		Version:     keys[l-1],
	}
}

func GetServiceAppKey(domainProject, env, appID string) string {
	return util.StringJoin([]string{
		GetServiceIndexRootKey(domainProject),
		env,
		appID,
	}, SPLIT)
}

func GenerateConsumerDependencyQueueKey(domainProject, consumerID, uuid string) string {
	return util.StringJoin([]string{
		GetServiceDependencyQueueRootKey(domainProject),
		consumerID,
		uuid,
	}, SPLIT)
}

func GetServiceDependencyQueueRootKey(domainProject string) string {
	return util.StringJoin([]string{
		GetRootKey(),
		RegistryServiceKey,
		RegistryDepsQueueKey,
		domainProject,
	}, SPLIT)
}
