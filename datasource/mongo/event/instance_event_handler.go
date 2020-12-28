/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package event

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/datasource/mongo"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/notify"
	"github.com/apache/servicecomb-service-center/server/syncernotify"
	"github.com/apache/servicecomb-service-center/syncer/samples/multi-servicecenters/servicecenter/hello-server/servicecenter"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v2"
)

// InstanceEventHandler is the handler to handle events
//as instance registry or instance delete, and notify syncer
type InstanceEventHandler struct {
	
}

func (h InstanceEventHandler) Type() string {
	return mongo.CollectionInstance
}

func (h InstanceEventHandler) OnEvent(evt sd.MongoEvent) {
	action := evt.Type
	instance := evt.Value.(sd.Instance)
	providerID := instance.InstanceInfo.ServiceId
	providerInstanceID := instance.InstanceInfo.InstanceId

	cacheService := sd.Store().Service().Cache().Get(providerID)
	var ms *discovery.MicroService
	if cacheService != nil {
		ms = cacheService.(sd.Service).ServiceInfo
	}
	if ms == nil {
		log.Info("get cached service failed, then get from data base")
		ms = getServiceFromDB(providerID) // service in the cache may not ready, query from db one time
		if ms == nil {
			log.Warn(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, get provider's file failed from db\n",
				action, providerID, providerInstanceID, instance.InstanceInfo.Endpoints))
			return
		}
	}
	if !syncernotify.GetSyncerNotifyCenter().Closed() {
		NotifySyncerInstanceEvent(evt, ms)
	}

	if notify.Center().Closed() {
		log.Warn(fmt.Sprintf("caught [%s] instance[%s/%s] event, endpoints %v, but notify service is closed\n",
			evt.Type, providerID, providerInstanceID, instance.InstanceInfo.Endpoints))
		return
	}
}

func NewInstanceEventHandler() *InstanceEventHandler {
	return &InstanceEventHandler{}
}

func NotifySyncerInstanceEvent(evt sd.MongoEvent, ms *discovery.MicroService) {
	instance := evt.Value.(sd.Instance).InstanceInfo
	log.Info(fmt.Sprintf("instance in NotifySyncerInstanceEvent : %v", instance))
	instanceKey := path.GenerateInstanceKey(evt.Value.(sd.Instance).Domain+"/"+
		evt.Value.(sd.Instance).Project, instance.ServiceId, instance.InstanceId)

	instanceKv := dump.KV{
		Key:   instanceKey,
		Value: instance,
	}

	dInstance := dump.Instance{
		KV:    &instanceKv,
		Value: instance,
	}
	serviceKey := path.GenerateServiceKey(evt.Value.(sd.Instance).Domain+"/"+
		evt.Value.(sd.Instance).Project, ms.ServiceId)
	serviceKv := dump.KV{
		Key:   serviceKey,
		Value: ms,
	}

	dService := dump.Microservice{
		KV:    &serviceKv,
		Value: ms,
	}

	instEvent := &dump.WatchInstanceChangedEvent{
		Action:   string(evt.Type),
		Service:  &dService,
		Instance: &dInstance,
	}
	syncernotify.GetSyncerNotifyCenter().AddEvent(instEvent)

	log.Debug(fmt.Sprintf("success to add instance change event:%s to event queue", instEvent))
}


func getServiceFromDB(providerID string) *discovery.MicroService {
	uri, err := getUri()
	if err != nil {
		log.Error("get uri from config err ", err)
		return nil
	}
	cfg := storage.NewConfig(uri)
	cfg.Timeout = "1s"
	service, errs := mongo.GetService(context.Background(), bson.M{"serviceinfo.serviceid": providerID})
	if errs != nil || service == nil {
		log.Error("get service from db err", err)
		return nil
	}
	return service.ServiceInfo
}

func getUri() (string, error) {
	var c *servicecenter.Config
	filePath := filepath.Join(util.GetAppRoot(), "conf", "app.yaml")
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error("get yaml file err", err)
		return "", err
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Error("Unmarshal yaml err", err)
		return "", err
	}
	return c.Registry.MongoDB.Cluster.Uri, nil
}
