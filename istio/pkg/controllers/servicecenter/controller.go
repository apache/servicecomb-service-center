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

package servicecenter

import (
	"reflect"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/event"
	"github.com/apache/servicecomb-service-center/istio/pkg/utils"
	"github.com/go-chassis/cari/discovery"

	client "github.com/apache/servicecomb-service-center/istio/pkg/controllers/servicecenter/client"
	"istio.io/pkg/log"
)

type Controller struct {
	// servicecomb service center go-chassis API client
	client *client.Client
	// Channel used to send and receive servicecomb service center change events from the service center controller
	event chan []event.ChangeEvent
	// Lock for service cache
	cacheMutex sync.Mutex
	// Cache of retrieved servicecomb service center microservices, mapped to their service ids
	serviceCache map[string]*event.MicroserviceEntry
}

func NewController(addr string, e chan []event.ChangeEvent) *Controller {
	controller := &Controller{
		client:       client.New(addr),
		event:        e,
		serviceCache: map[string]*event.MicroserviceEntry{},
	}

	return controller
}

// Run until a stop signal is received
func (c *Controller) Run(stop <-chan struct{}) {
	// start a new go routine to watch service center update
	go c.watchServiceCenter(stop)
}

// Stop the controller.
func (c *Controller) Stop() {
	// Unregister app instance watcher services
	for _, id := range c.client.AppInstanceWatcherCache {
		c.client.UnregisterInstanceWatcher(id)
	}
}

// Watch the service center registry for MicroService changes.
func (c *Controller) watchServiceCenter(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
			// Full sync all services from service center
			services, err := c.client.GetAllServices()
			if err != nil {
				log.Errorf("Failed to retrieve service center services from registry: err[%v]\n", err)
			}
			// Process received services
			c.onServiceCenterUpdate(services)
			// Number of seconds to wait between syncs
			time.Sleep(utils.PULL_INTERVAL)
		}
	}
}

// Send all new services to Istio controller.
func (c *Controller) onServiceCenterUpdate(services []*discovery.MicroService) {
	svcs := c.getChangedServices(services)
	if len(svcs) > 0 {
		c.event <- svcs
	}
}

// Send all instance events to Istio controller.
func (c *Controller) onInstanceUpdate(e event.ChangeEvent) {
	log.Debugf("New Service center instance event received from watcher: %+v\n", e)
	c.event <- []event.ChangeEvent{e}
}

// Get service center service changes, register watcher for newly created services.
func (c *Controller) getChangedServices(services []*discovery.MicroService) []event.ChangeEvent {
	// All non-watcher service center services mapped to their ids
	currServices := map[string]*event.MicroserviceEntry{}
	// All new non-watcher service center services
	newServices := []*event.MicroserviceEntry{}
	// IDs of current service center watcher services
	currAppInstanceWatcherIds := map[string]string{}
	// Service events that must be pushed
	changes := []event.ChangeEvent{}
	for _, s := range services {
		name := s.ServiceName
		appId := s.AppId
		id := s.ServiceId
		if name != utils.WATCHER_SVC_NAME && name != utils.SERVICECENTER_ETCD_NAME && name != utils.SERVICECENTER_MONGO_NAME {
			entry := &event.MicroserviceEntry{MicroService: s}
			if cachedEntry, ok := c.serviceCache[id]; !ok {
				if _, ok := c.client.AppInstanceWatcherCache[appId]; !ok {
					// Register new app instance watcher service
					watcherId, err := c.client.RegisterAppInstanceWatcher(utils.WATCHER_SVC_NAME, appId, c.onInstanceUpdate)
					if err != nil {
						continue
					}
					// Record the id of the watcher service for this app
					currAppInstanceWatcherIds[appId] = watcherId
				}
				// Collect newly created service
				changeEvent := event.ChangeEvent{Action: discovery.EVT_CREATE, Event: entry}
				changes = append(changes, changeEvent)
				newServices = append(newServices, entry)
				currServices[id] = entry
			} else {
				if !reflect.DeepEqual(s, cachedEntry.MicroService) {
					// Collect updated service
					changeEvent := event.ChangeEvent{Action: discovery.EVT_UPDATE, Event: entry}
					changes = append(changes, changeEvent)
					currServices[id] = entry
				} else {
					// No change, keep cache entry
					currServices[id] = cachedEntry
				}
			}
		} else if name == utils.WATCHER_SVC_NAME && c.client.AppInstanceWatcherCache[appId] == id {
			// Watcher still exists as expected, record its current id
			currAppInstanceWatcherIds[appId] = id
		}
	}
	// Collect deleted services
	for id, cachedEntry := range c.serviceCache {
		if _, ok := currServices[id]; !ok {
			changes = append(changes, event.ChangeEvent{
				Action: discovery.EVT_DELETE,
				Event:  cachedEntry,
			})
		}
	}
	// Initial sync-up for newly created services; retrieve and start watching their instances
	c.initNewServices(newServices)
	// Update service ID cache with current services
	c.refreshServiceCache(currServices)
	// Check for app instance watcher changes
	c.checkAppInstanceWatchers(currAppInstanceWatcherIds)

	return changes
}

// Save MicroService(s) retrieved from service center registry
func (c *Controller) refreshServiceCache(services map[string]*event.MicroserviceEntry) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.serviceCache = services
}

// Detect missing watcher services in registry. If a watcher service was expected but is missing, flag it to be re-registered.
func (c *Controller) checkAppInstanceWatchers(currAppInstanceWatcherIds map[string]string) {
	for appId := range c.client.AppInstanceWatcherCache {
		if _, ok := currAppInstanceWatcherIds[appId]; !ok {
			log.Warnf("Instance watcher for appId %s is invalid, invalidating its cache entries", appId)
			// Watcher is missing for this app, remove all app's services from cache
			newServiceCache := map[string]*event.MicroserviceEntry{}
			for id, key := range c.serviceCache {
				if key.MicroService.AppId != appId {
					newServiceCache[id] = key
				}
			}
			c.refreshServiceCache(newServiceCache)
		}
	}
	// Cache current watcher ids (if any are missing, will be re-registered on next sync)
	c.client.RefreshAppInstanceWatcherCache(currAppInstanceWatcherIds)
}

// Watch services, has side effect of adding instances to MicroserviceEntry(s)
func (c *Controller) initNewServices(newServices []*event.MicroserviceEntry) {
	serviceInstanceMap := c.client.GetServiceInstances(newServices)
	if serviceInstanceMap == nil {
		return
	}
	for _, s := range newServices {
		if instances, ok := serviceInstanceMap[s.MicroService.ServiceId]; ok {
			for _, instance := range instances {
				s.Instances = append(s.Instances, &event.InstanceEntry{
					MicroServiceInstance: instance,
				})
			}
		}
	}
}
