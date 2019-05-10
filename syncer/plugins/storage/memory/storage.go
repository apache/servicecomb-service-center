package memory

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"

	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const PluginName = "memory"

var (
	defaultMapping = make(pb.SyncMapping)
	snapshotPath   = "./data/syncer-snapshot"
)

func init() {
	// Register self as a storage plugin
	plugins.RegisterPlugin(&plugins.Plugin{
		Kind: plugins.PluginStorage,
		Name: PluginName,
		New:  New,
	})
}

func New() plugins.PluginInstance {
	return &Repo{
		syncData:    &pb.SyncData{},
		intsMapping: loadSnapshot(),
	}
}

// Repo Repository struct
type Repo struct {
	syncData *pb.SyncData

	// mapping table for other datacenter instances
	intsMapping map[string]pb.SyncMapping
	lock        sync.RWMutex
}

// loadSnapshot Load snapshot of mapping table
func loadSnapshot() map[string]pb.SyncMapping {
	mapping := make(map[string]pb.SyncMapping)
	data, err := ioutil.ReadFile(snapshotPath)
	if err != nil {
		log.Warnf("get syncer snapshot from '%s' failed, error: %s", snapshotPath, err)
		return mapping
	}
	err = json.Unmarshal(data, mapping)
	if err != nil {
		log.Warnf("unmarshal syncer snapshot failed, error: %s", err)
	}
	return mapping
}

func (r *Repo) Stop() {
	r.flush()
}

// flush Refresh the mapping table to the hard disk
func (r *Repo) flush() {
	data, err := json.Marshal(r.intsMapping)
	if err != nil {
		log.Warnf("marshal syncer snapshot failed, error: %s", err)
		return
	}

	f, err := utils.OpenFile(snapshotPath)
	if err != nil {
		log.Warnf("open syncer snapshot file '%s' failed, error: %s", snapshotPath, err)
		return
	}

	_, err = f.Write(data)
	if err != nil {
		log.Warnf("flush syncer snapshot to '%s' failed, error: %s", snapshotPath, err)
		return
	}
}

// SaveSyncData Save self sync data
func (r *Repo) SaveSyncData(data *pb.SyncData) {
	r.lock.Lock()
	r.syncData = data
	r.lock.Unlock()
}

// GetSyncData Get self sync data
func (r *Repo) GetSyncData() (data *pb.SyncData) {
	r.lock.RLock()
	data = &pb.SyncData{Services: r.syncData.Services[:]}
	r.lock.RUnlock()
	return
}

// SaveSyncMapping Save mapping table for other datacenter instances
func (r *Repo) SaveSyncMapping(nodeName string, mapping pb.SyncMapping) {
	r.lock.Lock()
	r.intsMapping[nodeName] = mapping
	r.lock.Unlock()
}

// GetSyncMapping Get mapping table for other datacenter instances
func (r *Repo) GetSyncMapping(nodeName string) (mapping pb.SyncMapping) {
	r.lock.RLock()
	data, ok := r.intsMapping[nodeName]
	if !ok {
		data = defaultMapping
	}
	r.lock.RUnlock()
	return data
}

// GetAllMapping Get all mapping table for other datacenters instances
func (r *Repo) GetAllMapping() (mapping pb.SyncMapping) {
	r.lock.RLock()
	mapping = make(pb.SyncMapping)
	for _, data := range r.intsMapping {
		if data != nil {
			for key, val := range data {
				mapping[key] = val
			}
		}
	}
	r.lock.RUnlock()
	return
}
