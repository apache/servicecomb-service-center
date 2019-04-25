package plugins

import (
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository"
	"github.com/apache/servicecomb-service-center/syncer/plugins/storage"
)

type PluginType int

const (
	PluginStorage PluginType = iota
	PluginRepository
	pluginTotal

	BUILDIN       = "buildin"
	STATIC        = "static"
	DYNAMIC       = "dynamic"
	keyPluginName = "name"
)

func (p PluginType) String() string {
	switch p {
	case PluginStorage:
		return "storage"
	case PluginRepository:
		return "repository"
	default:
		return ""
	}
}

func (m Manager) Storage() storage.Repository { return m.Instance(PluginStorage).(storage.Repository) }
func (m Manager) Repository() repository.Adaptor {
	return m.Instance(PluginRepository).(repository.Adaptor)
}
