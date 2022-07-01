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

package client

import (
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/istio/pkg/event"

	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	"istio.io/pkg/log"
)

const (
	// Time in seconds to wait before re-registering a deleted Servicecomb Service Center Watcher service
	REREGISTER_INTERVAL time.Duration = time.Second * 5
)

// Servicecomb Service Center go-chassis client
type Client struct {
	client                  *sc.Client
	cacheMutex              sync.Mutex
	AppInstanceWatcherCache map[string]string // Maps appId to id of a instance watcher service. Need app-specific watchers to avoid cross-app errors.
}

func New(addr string) *Client {
	registryClient, err := sc.NewClient(
		sc.Options{
			Endpoints: []string{addr},
		})
	if err != nil {
		log.Errorf("Failed to create service center client, err[%v]\n", err)
	}
	return &Client{
		client:                  registryClient,
		AppInstanceWatcherCache: map[string]string{},
	}
}

// Check whether a service center MicroService exists in the registry.
func (c *Client) GetServiceExistence(microServiceId string) bool {
	s, _ := c.client.GetMicroService(microServiceId, sc.WithGlobal())
	return s != nil
}

// Retrieve all service center MicroServices, without their instances, from the registry.
func (c *Client) GetAllServices() ([]*discovery.MicroService, error) {
	microservices, err := c.client.GetAllMicroServices(sc.WithGlobal())
	if err != nil {
		return nil, err
	}
	return microservices, err
}

// Register a new service center Watcher service that watches instance-level change events for all service center services sharing a specific appId.
func (c *Client) RegisterAppInstanceWatcher(name string, appId string, callback func(event event.ChangeEvent)) (string, error) {
	watcherService := &discovery.MicroService{
		AppId:       appId,
		ServiceName: name,
		Environment: "",
		Version:     "0.0.1",
	}

	prevId, _ := c.client.GetMicroServiceID(watcherService.AppId, watcherService.ServiceName, "0.0.1", watcherService.Environment, sc.WithGlobal())
	if prevId != "" {
		// Need to reregister existing watcher for this app to reestablish the websocket connection
		log.Warnf("Instance watcher already exists in service center registry with id %s, attempting to unregister...\n", prevId)
		err := c.UnregisterInstanceWatcher(prevId)
		if err != nil {
			log.Errorf("Failed to unregister exising instance watcher in service center registry with id %s, err[%v]\n", prevId, err)
			return "", err
		}
		log.Infof("Successfully unregistered instance watcher, sleeping for %s before re-registering...\n", REREGISTER_INTERVAL)
		// Sleep allows time for service center to unregister watcher consumer/producer relationships
		// Re-registering too early will cause race conditions
		time.Sleep(REREGISTER_INTERVAL)
	}
	id, err := c.client.RegisterService(watcherService)
	if err != nil {
		log.Errorf("Failed to register instance watcher in service center registry with service name %s, err[%v]\n", name, err)
		return "", err
	}
	err = c.client.WatchMicroService(id, func(e *sc.MicroServiceInstanceChangedEvent) {
		callback(event.ChangeEvent{Action: discovery.EventType(e.Action), Event: &event.InstanceEntry{MicroServiceInstance: e.Instance}})
	})
	if err != nil {
		log.Errorf("Failed to watch service center instances using watcher service %s, err[%v]\n", name, err)
	}
	// Cache the id of the app instance watcher service
	c.AppInstanceWatcherCache[appId] = id
	log.Debugf("Registered instance watcher with service name %s and id %s for appId %s\n", name, id, appId)
	return id, nil
}

// Unregister a service center Watcher service.
func (c *Client) UnregisterInstanceWatcher(serviceId string) error {
	if !c.GetServiceExistence(serviceId) {
		log.Debug("Instance watcher no longer exists in registry, skipping unregister...")
		return nil
	}
	if serviceId != "" {
		ok, err := c.client.UnregisterMicroService(serviceId)
		if !ok {
			log.Warnf("failed to unregister instance watcher in service center registry with service name %s, err[%v]\n", serviceId, err)
		} else {
			log.Debug("Instance watcher successfully unregistered")
		}
		return err
	} else {
		return nil
	}
}

// GetServiceInstances fetch newly received service instances.
func (c *Client) GetServiceInstances(entries []*event.MicroserviceEntry) map[string][]*discovery.MicroServiceInstance {
	if len(entries) == 0 {
		return nil
	}
	appServiceKeys := make(map[string][]*discovery.FindService, len(entries))
	for _, e := range entries {
		s := e.MicroService
		appId := s.AppId
		key := &discovery.MicroServiceKey{
			ServiceName: s.ServiceName,
			AppId:       appId,
			Environment: s.Environment,
			Version:     s.Version,
			Alias:       s.Alias,
		}
		if _, ok := c.AppInstanceWatcherCache[appId]; !ok {
			log.Errorf("Failed to watch new microservices for appId %s, watcher service failed to start\n", appId)
			continue
		}
		appServiceKeys[appId] = append(appServiceKeys[appId], &discovery.FindService{Service: key})
	}
	serviceInstanceMap := map[string][]*discovery.MicroServiceInstance{}
	for appId, keys := range appServiceKeys {
		// Initial instance sync of services with same appId, will use app instance watcher for future instance updates
		res, err := c.client.BatchFindInstances(c.AppInstanceWatcherCache[appId], keys, sc.WithGlobal(), sc.WithoutRevision())
		if err != nil {
			log.Errorf("Failed to watch new microservices, unable to get instances, err[%v]\n", err)
		}
		for _, r := range res.Services.Updated {
			e := entries[r.Index]
			serviceInstanceMap[e.MicroService.ServiceId] = r.Instances
		}
		log.Infof("Started watching instances of services with appId %s\n", appId)
	}
	for _, e := range entries {
		if _, ok := serviceInstanceMap[e.MicroService.ServiceId]; !ok {
			log.Errorf("Failed to watch instances of service %s with id %s\n", e.MicroService.ServiceName, e.MicroService.ServiceId)
		}
	}
	return serviceInstanceMap
}

// Save the ids of currently active service center Watcher services mapped to the appIds that they are responsible for.
func (c *Client) RefreshAppInstanceWatcherCache(appWatcherIds map[string]string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.AppInstanceWatcherCache = appWatcherIds
}
