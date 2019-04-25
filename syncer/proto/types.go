package proto

type SyncMapping map[string]*SyncServiceKey

type SyncServiceKey struct {
	DomainProject string `json:"domain_project"`
	ServiceID     string `json:"service_id"`
	InstanceID    string `json:"instance_id"`
}

type NodeDataInfo struct {
	NodeName string
	DataInfo *SyncData
}

type ServiceKey struct {
	ServiceName string
	Version     string
}
