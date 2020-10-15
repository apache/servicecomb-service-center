package etcd

import "github.com/apache/servicecomb-service-center/pkg/util"

const (
	RegistryRootKey    = "cse-sr"
	RegistryProjectKey = "projects"
	SPLIT              = "/"
	RegistryDomainKey  = "domains"
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
