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
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/event"
	"github.com/apache/servicecomb-service-center/istio/pkg/utils"
	"github.com/go-chassis/cari/discovery"

	"istio.io/pkg/log"
)

type Controller struct {
	// servicecomb service center go-chassis API client
	conn *Connector
	// Channel used to send and receive servicecomb service center change events from the service center controller
	events chan []event.ChangeEvent
	// Cache of retrieved servicecomb service center microservices, mapped to their service ids
	serviceCache sync.Map
}

func NewController(addr string, e chan []event.ChangeEvent) *Controller {
	controller := &Controller{
		conn:         NewConnector(addr),
		events:       e,
		serviceCache: sync.Map{},
	}

	return controller
}

// Run until a stop signal is received
func (c *Controller) Run(ctx context.Context) {
	// start a new go routine to watch service center update
	go c.watchServiceCenter(ctx)
}

// Stop the controller.
func (c *Controller) Stop() {
	// Unregister app instance watcher services
	c.conn.AppInstanceWatcherCache.Range(func(_, value interface{}) bool {
		c.conn.UnregisterInstanceWatcher(value.(string))
		return true
	})
}

// Watch the service center registry for MicroService changes.
func (c *Controller) watchServiceCenter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Full sync all services from service center
			services, err := c.conn.GetAllServices()
			if err != nil {
				log.Errorf("failed to retrieve service center services from registry: err[%v]\n", err)
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
		c.events <- svcs
	}
}

// Send all instance events to Istio controller.
func (c *Controller) onInstanceUpdate(e event.ChangeEvent) {
	log.Debugf("new service center instance event received from watcher: %+v\n", e)
	c.events <- []event.ChangeEvent{e}
}

// Get service center service changes, register watcher for newly created services.
func (c *Controller) getChangedServices(services []*discovery.MicroService) []event.ChangeEvent {
	// All non-watcher service center services mapped to their ids
	currServices := sync.Map{}
	// All new non-watcher service center services
	newServices := []*event.MicroserviceEntry{}
	// IDs of current service center watcher services
	currAppInstanceWatcherIds := sync.Map{}
	// Service events that must be pushed
	changes := []event.ChangeEvent{}
	for _, s := range services {
		name := s.ServiceName
		appId := s.AppId
		id := s.ServiceId
		if name != utils.WATCHER_SVC_NAME && name != utils.SERVICECENTER_ETCD_NAME && name != utils.SERVICECENTER_MONGO_NAME {
			entry := &event.MicroserviceEntry{MicroService: s}
			if cachedEntry, ok := c.serviceCache.Load(id); !ok {
				if _, ok := c.conn.AppInstanceWatcherCache.Load(appId); !ok {
					// Register new app instance watcher service
					watcherId, err := c.conn.RegisterAppInstanceWatcher(utils.WATCHER_SVC_NAME, appId, c.onInstanceUpdate)
					if err != nil {
						continue
					}
					// Record the id of the watcher service for this app
					currAppInstanceWatcherIds.Store(appId, watcherId)
				}
				// Collect newly created service
				changeEvent := event.ChangeEvent{Action: discovery.EVT_CREATE, Event: entry}
				changes = append(changes, changeEvent)
				newServices = append(newServices, entry)
				currServices.Store(id, entry)
			} else {
				cachedEntryEvent := cachedEntry.(*event.MicroserviceEntry)
				if !reflect.DeepEqual(s, cachedEntryEvent.MicroService) {
					// Collect updated service
					changeEvent := event.ChangeEvent{Action: discovery.EVT_UPDATE, Event: entry}
					changes = append(changes, changeEvent)
					currServices.Store(id, entry)
				} else {
					// No change, keep cache entry
					currServices.Store(id, cachedEntry)
				}
			}
		} else if name == utils.WATCHER_SVC_NAME {
			if k, ok := c.conn.AppInstanceWatcherCache.Load(appId); ok {
				if k.(string) == id {
					// Watcher still exists as expected, record its current id
					currAppInstanceWatcherIds.Store(appId, id)
				}
			}

		}
	}
	// Collect deleted services
	c.serviceCache.Range(func(key, value interface{}) bool {
		if _, ok := currServices.Load(key); !ok {
			changes = append(changes, event.ChangeEvent{
				Action: discovery.EVT_DELETE,
				Event:  value.(*event.MicroserviceEntry),
			})
		}
		return true
	})

	// Initial sync-up for newly created services; retrieve and start watching their instances
	c.initNewServices(newServices)
	// Update service ID cache with current services
	c.refreshServiceCache(currServices)
	// Check for app instance watcher changes
	c.checkAppInstanceWatchers(currAppInstanceWatcherIds)

	return changes
}

// Save MicroService(s) retrieved from service center registry
func (c *Controller) refreshServiceCache(services sync.Map) {
	c.serviceCache = services
}

// Detect missing watcher services in registry. If a watcher service was expected but is missing, flag it to be re-registered.
func (c *Controller) checkAppInstanceWatchers(currAppInstanceWatcherIds sync.Map) {
	c.conn.AppInstanceWatcherCache.Range(func(appId, _ interface{}) bool {
		if _, ok := currAppInstanceWatcherIds.Load(appId); !ok {
			log.Warnf("instance watcher for appId %s is invalid, invalidating its cache entries", appId)
			// Watcher is missing for this app, remove all app's services from cache
			newServiceCache := sync.Map{}
			c.serviceCache.Range(func(key, value interface{}) bool {
				microServiceValue := value.(*event.MicroserviceEntry)
				if microServiceValue.MicroService.AppId != appId {
					newServiceCache.Store(key, value)
				}
				return true
			})

			c.refreshServiceCache(newServiceCache)
		}
		return true
	})
	// Cache current watcher ids (if any are missing, will be re-registered on next sync)
	c.conn.RefreshAppInstanceWatcherCache(currAppInstanceWatcherIds)
}

// Watch services, has side effect of adding instances to MicroserviceEntry(s)
func (c *Controller) initNewServices(newServices []*event.MicroserviceEntry) {
	serviceInstanceMap := c.conn.GetServiceInstances(newServices)
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
